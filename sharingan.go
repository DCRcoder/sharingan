package sharingan

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/panjf2000/ants"
)

var defaultPoolSize = 50

// Serialization 序列化接口
type Serialization interface {
	Serialize(ctx context.Context, src interface{}) error
	SetSharingan(s *Sharingan)
}

// Sharingan  序列化
type Sharingan struct {
	prepareData sync.Map
	wg          sync.WaitGroup
	isAsync     bool
	poolSize    int
}

// NewSharingan new instance
func NewSharingan() *Sharingan {
	return &Sharingan{
		prepareData: sync.Map{},
		wg:          sync.WaitGroup{},
		isAsync:     false,
		poolSize:    defaultPoolSize,
	}
}

// SetPrepare 设置序列化需要的预存数据
func (d *Sharingan) SetPrepare(prepare map[string]interface{}) {
	for k, v := range prepare {
		d.prepareData.Store(k, v)
	}
}

// GetPrapareData 获取预存数据
func (d *Sharingan) GetPrapareData(key string) (interface{}, bool) {
	return d.prepareData.Load(key)
}

// Converted 单个序列化
func (d *Sharingan) Converted(ctx context.Context, src interface{}, dest Serialization) error {
	err := dest.Serialize(ctx, src)
	if d.isAsync {
		d.wg.Done()
	}
	return err
}

// ConvertedMany many 序列化 同步异步可选
func (d *Sharingan) ConvertedMany(ctx context.Context, src interface{}, dest interface{}, async bool) error {
	srcValue := reflect.ValueOf(src)
	if srcValue.Type().Kind() != reflect.Slice {
		return fmt.Errorf("ConvertedMany src must be slice type")
	}
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	if !destValue.CanSet() {
		return fmt.Errorf("ConvertedMany dest must be pointer")
	}
	d.isAsync = async
	if async {
		return d.asyncConvert(ctx, srcValue, destValue)
	} else {
		return d.syncConvert(ctx, srcValue, destValue)
	}
}

// syncConvert 同步转化
func (d *Sharingan) syncConvert(ctx context.Context, srcValue reflect.Value, destValue reflect.Value) error {
	dests := reflect.MakeSlice(destValue.Type(), srcValue.Len(), srcValue.Len())
	for i := 0; i < srcValue.Len(); i++ {
		var si interface{}
		value := srcValue.Index(i)
		if value.Type().Kind() != reflect.Ptr {
			return fmt.Errorf("src element must be pointer")
		}
		si = value.Interface()
		ds := dests.Index(i)
		y := reflect.New(ds.Type().Elem())
		switch y.Type().Kind() {
		case reflect.Ptr:
			r := y.Interface().(Serialization)
			r.SetSharingan(d)
			d.Converted(ctx, si, r)
			ds.Set(reflect.ValueOf(r))
		default:
			return fmt.Errorf("dest element must be pointer")
		}
	}
	destValue.Set(dests)
	return nil
}

// asyncConvert 异步转化
func (d *Sharingan) asyncConvert(ctx context.Context, srcValue reflect.Value, destValue reflect.Value) error {
	pool, _ := ants.NewPool(d.poolSize)
	dests := reflect.MakeSlice(destValue.Type(), srcValue.Len(), srcValue.Len())
	for i := 0; i < srcValue.Len(); i++ {
		var si interface{}
		value := srcValue.Index(i)
		if value.Type().Kind() != reflect.Ptr {
			return fmt.Errorf("src element must be pointer")
		}
		si = value.Interface()
		ds := dests.Index(i)
		y := reflect.New(ds.Type().Elem())
		switch y.Type().Kind() {
		case reflect.Ptr:
			r := y.Interface().(Serialization)
			r.SetSharingan(d)
			d.wg.Add(1)
			pool.Submit(func() { _ = d.Converted(ctx, si, r) })
			ds.Set(reflect.ValueOf(r))
		default:
			return fmt.Errorf("dest element must be pointer")
		}
	}
	d.wg.Wait()
	destValue.Set(dests)
	return nil
}
