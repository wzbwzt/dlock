package dlock

import "time"

type IRegister interface {
	Lock(string, time.Duration) error
	Unlock(string) error
}

type Dlock struct {
	//注册类型
	register IRegister
}

type Option func(*Dlock)

func NewDlock(opts ...Option) *Dlock {
	lock := &Dlock{}
	for _, opt := range opts {
		opt(lock)
	}
	return lock
}

func WithRegister(i IRegister) Option {
	return func(d *Dlock) {
		d.register = i
	}
}

func (d *Dlock) Lock(path string, duration time.Duration) error {
	d.register.Lock(path, duration)
	return nil
}

func (d *Dlock) UnLock(path string) error {
	d.register.Unlock(path)
	return nil
}
