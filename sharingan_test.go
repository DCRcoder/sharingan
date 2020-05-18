package sharingan

import (
	"context"
	"testing"
	"fmt"
	"encoding/json"
)

type Person struct {
	ID   string
	Name string
}

type People struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Sex  string `json:"sex"`
	*Sharingan
}


func (b *People) Serialize(ctx context.Context, src interface{}) error {
	ori := src.(*Person)
	b.ID = ori.ID
	b.Name = ori.Name
	sex, ok := b.GetPrapareData("sex")
	if ok {
		b.Sex = sex.(string)
	} else {
		b.Sex = "gay"
	}
	return nil
}

func (b *People) SetSharingan(s *Sharingan) {
	b.Sharingan = s
}

func TestSerializeWithoutPrepareData(t *testing.T) {
	ps := &Person{ID: "b", Name: "s"}
	pe := &People{}
	sgan := NewSharingan()
	pe.SetSharingan(sgan)

	sgan.Converted(context.TODO(), ps, pe)
	c, _ := json.Marshal(pe)
	fmt.Println(string(c))
	if string(c) != `{"id":"b","name":"s","sex":"gay"}` {
		t.Errorf("[TestSerialize]%v\n", string(c))
	}
}

func TestSerializeWithPrepareData(t *testing.T) {
	ps := &Person{ID: "b", Name: "s"}
	pe := &People{}
	sgan := NewSharingan()
	pe.SetSharingan(sgan)
	z := map[string]interface{}{
		"sex": "man",
	}
	sgan.SetPrepare(z)
	sgan.Converted(context.TODO(), ps, pe)
	c, _ := json.Marshal(pe)
	fmt.Println(string(c))
	if string(c) != `{"id":"b","name":"s","sex":"man"}` {
		t.Errorf("[TestSerialize]%v\n", string(c))
	}
}

func TestSerializeManyAsync(t *testing.T) {
	s := &Person{ID: "b", Name: "s"}
	s1 := &Person{ID: "b1", Name: "s1"}
	sgan := NewSharingan()
	a := make([]*Person, 0)
	a = append(a, s)
	a = append(a, s1)
	bz := make([]*People, 100)
	err := sgan.ConvertedMany(context.TODO(), a, &bz, true)
	if len(bz) != 2 {
		t.Errorf("[TestSerializeMany] convert data error data:%v, result:%v",a, bz)
	}
	if bz[1].ID != "b1" || bz[1].Name != "s1" || bz[1].Sex != "gay" {
		t.Errorf("[TestSerializeMany] convert data error data:%v, result:%v",a[1], bz[1])
	}
	if err != nil {
		t.Errorf("[TestSerializeMany] convert error %v", err)
	}
}

func TestSerializeManySync(t *testing.T) {
	s := &Person{ID: "b", Name: "s"}
	s1 := &Person{ID: "b1", Name: "s1"}
	sgan := NewSharingan()
	z := map[string]interface{}{
		"sex": "man",
	}
	sgan.SetPrepare(z)
	a := make([]*Person, 0)
	a = append(a, s)
	a = append(a, s1)
	bz := make([]*People, 100)
	err := sgan.ConvertedMany(context.TODO(), a, &bz, true)
	if len(bz) != 2 {
		t.Errorf("[TestSerializeMany] convert data error data:%v, result:%v",a, bz)
	}
	if bz[0].ID != "b" || bz[0].Name != "s" || bz[0].Sex != "man" {
		t.Errorf("[TestSerializeMany] convert data error data:%v, result:%v",a[0], bz[0])
	}
	if err != nil {
		t.Errorf("[TestSerializeMany] convert error %v", err)
	}
}