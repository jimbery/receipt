package match

import (
	"sort"
	"time"

	"github.com/jimbery/receipt/internal/model"
)

const amountBucketMinor = 500 // £5 buckets for candidate indexing

type receiptIndex struct {
	byCurrency map[string][]indexedReceipt
}

type indexedReceipt struct {
	receipt model.Receipt
	bucket  int64
	time    time.Time
}

func buildReceiptIndex(receipts []model.Receipt, skip map[string]struct{}) receiptIndex {
	idx := receiptIndex{byCurrency: make(map[string][]indexedReceipt)}
	for _, r := range receipts {
		if _, skip := skip[r.ID]; skip {
			continue
		}
		cur := r.Total.Currency
		idx.byCurrency[cur] = append(idx.byCurrency[cur], indexedReceipt{
			receipt: r,
			bucket:  r.Total.Amount / amountBucketMinor,
			time:    r.IssuedAt.UTC,
		})
	}
	for cur := range idx.byCurrency {
		sort.Slice(idx.byCurrency[cur], func(i, j int) bool {
			return idx.byCurrency[cur][i].time.Before(idx.byCurrency[cur][j].time)
		})
	}
	return idx
}

func (e *Engine) generateCandidates(
	transactions []model.Transaction,
	receipts []model.Receipt,
	skipTxn, skipReceipt map[string]struct{},
) []candidate {
	idx := buildReceiptIndex(receipts, skipReceipt)
	looseAbs := e.cfg.AbsAmountTolerance * candidateAbsToleranceMultiplier
	looseRel := e.cfg.RelAmountTolerance * candidateRelToleranceMultiplier

	var out []candidate
	for _, txn := range transactions {
		if _, skip := skipTxn[txn.ID]; skip {
			continue
		}
		out = append(out, e.candidatesForTransaction(txn, idx, looseAbs, looseRel)...)
	}
	return out
}

func (e *Engine) candidatesForTransaction(
	txn model.Transaction,
	idx receiptIndex,
	looseAbs int64,
	looseRel float64,
) []candidate {
	list := idx.byCurrency[txn.Amount.Currency]
	if len(list) == 0 {
		return nil
	}

	txnBucket := txn.Amount.Amount / amountBucketMinor
	fuelRel := fuelPreAuthRel(txn, e.cfg)
	bucketSpan := amountBucketSpan(txn.Amount.Amount, looseAbs, looseRel, fuelRel)
	pairLooseRel := pairAmountTolerance(looseRel, fuelRel)

	var out []candidate
	for _, ir := range list {
		if !bucketMatches(txnBucket, ir.bucket, bucketSpan) {
			continue
		}
		if !e.withinTemporalWindow(txn.OccurredAt.UTC, ir.time) {
			continue
		}
		if !txn.Amount.WithinTolerance(ir.receipt.Total, looseAbs, pairLooseRel) {
			continue
		}
		c := candidate{Transaction: txn, Receipt: ir.receipt}
		if e.cfg.MinMerchantForCandidate > 0 {
			_, signals := e.scoreCandidate(c)
			if signals["merchant"] < e.cfg.MinMerchantForCandidate {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func fuelPreAuthRel(txn model.Transaction, cfg Config) float64 {
	if txn.MCC == "5541" {
		return cfg.AmountBands.FuelPreAuthRel
	}
	return 0
}

func pairAmountTolerance(looseRel, fuelRel float64) float64 {
	if fuelRel > looseRel {
		return fuelRel
	}
	return looseRel
}

func bucketMatches(txnBucket, receiptBucket, span int64) bool {
	return receiptBucket >= txnBucket-span && receiptBucket <= txnBucket+span
}

func amountBucketSpan(txnAmount, looseAbs int64, looseRel, fuelPreAuthRel float64) int64 {
	span := looseAbs/amountBucketMinor + 1
	maxRel := looseRel
	if fuelPreAuthRel > maxRel {
		maxRel = fuelPreAuthRel
	}
	if relSpan := int64(float64(txnAmount) * maxRel / float64(amountBucketMinor)); relSpan+1 > span {
		span = relSpan + 1
	}
	return span
}
