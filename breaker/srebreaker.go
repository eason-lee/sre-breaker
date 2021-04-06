package breaker

import (
	"math"
	"sre-breaker/breaker/utils"
	"time"
)

const (
	// 250ms for bucket duration
	window     = time.Second * 10
	buckets    = 40
	k          = 1.5
	protection = 5
)

type sreBreaker struct {
	k     float64              // 倍数值
	stat  *utils.RollingWindow // 滑动时间窗口，请求成功和失败的计数
	proba *utils.Proba         // 动态概率
}

func newSreBreaker() *sreBreaker {
	bucketDuration := time.Duration(int64(window) / int64(buckets))
	st := utils.NewRollingWindow(buckets, bucketDuration)
	return &sreBreaker{
		stat:  st,
		k:     k,
		proba: utils.NewProba(),
	}
}

func (b *sreBreaker) accept() error {
	accepts, total := b.history() // 接受的请求数和总请求数
	weightedAccepts := b.k * float64(accepts)
	dropRatio := math.Max(0, (float64(total-protection)-weightedAccepts)/float64(total+1))
	if dropRatio <= 0 {
		return nil
	}

	if b.proba.TrueOnProba(dropRatio) {
		return ErrServiceUnavailable
	}

	return nil
}

func (b *sreBreaker) allow() (internalPromise, error) {
	if err := b.accept(); err != nil {
		return nil, err
	}

	return srePromise{
		b: b,
	}, nil
}

func (b *sreBreaker) doReq(req func() error, fallback func(err error) error, acceptable Acceptable) error {
	if err := b.accept(); err != nil {
		if fallback != nil {
			return fallback(err)
		}

		return err
	}

	defer func() {
		if e := recover(); e != nil {
			b.markFailure()
			panic(e)
		}
	}()

	err := req()
	if acceptable(err) {
		b.markSuccess()
	} else {
		b.markFailure()
	}

	return err
}

func (b *sreBreaker) markSuccess() {
	b.stat.Add(1)
}

func (b *sreBreaker) markFailure() {
	b.stat.Add(0)
}

func (b *sreBreaker) history() (accepts int64, total int64) {
	b.stat.Reduce(func(b *utils.Bucket) {
		accepts += int64(b.Sum)
		total += b.Count
	})

	return
}

type srePromise struct {
	b *sreBreaker
}

func (p srePromise) Accept() {
	p.b.markSuccess()
}

func (p srePromise) Reject() {
	p.b.markFailure()
}
