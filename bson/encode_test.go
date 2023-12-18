package bson

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PStruct struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
}

type PStructPointer struct {
	ID    *primitive.ObjectID `bson:"_id,omitempty"`
	Regex *primitive.Regex    `bson:"regex,omitempty"`
}

type BStruct struct {
	ID ObjectId `bson:"_id,omitempty"`
}

type BStructPointer struct {
	ID    *ObjectId `bson:"_id,omitempty"`
	Regex *RegEx    `bson:"regex,omitempty"`
}

func TestMarshal_ObjectID(t *testing.T) {
	t.Run("real id", func(t *testing.T) {
		id := primitive.NewObjectID()

		p := PStruct{ID: id}
		b := BStruct{ID: ObjectIdHex(id.Hex())}

		CheckMarshalAndUnmarshal(t, p, b)
	})

	t.Run("zero id", func(t *testing.T) {
		p := PStruct{}
		b := BStruct{}

		CheckMarshalAndUnmarshal(t, p, b)
	})

	// this actually works but is hard to test.
	// pointer to empty value will be empty
	// t.Run("zero id pointer", func(t *testing.T) {
	// 	id := primitive.ObjectID{}
	// 	p := PStructPointer{ID: &id}

	// 	i := ObjectId("")
	// 	b := BStructPointer{ID: &i}

	// 	CheckMarshalAndUnmarshal(t, p, b)
	// })

	t.Run("not-zero pointer id", func(t *testing.T) {
		id := primitive.NewObjectID()
		p := PStructPointer{ID: &id}

		i := ObjectIdHex(id.Hex())
		b := BStructPointer{ID: &i}

		CheckMarshalAndUnmarshal(t, p, b)
	})

	t.Run("nil id", func(t *testing.T) {
		p := PStructPointer{ID: nil}
		b := BStructPointer{ID: nil}
		CheckMarshalAndUnmarshal(t, p, b)
	})

	t.Run("nil id into old non-pointer", func(t *testing.T) {
		type Struct struct {
			ID *primitive.ObjectID `bson:"_id"`
		}

		data, err := Marshal(Struct{ID: nil})
		assert.NoError(t, err)

		type Struct2 struct {
			ID ObjectId `bson:"_id"`
		}

		// no error, sets as empty
		{
			// old driver
			var s2 Struct2
			err = Unmarshal(data, &s2)
			assert.NoError(t, err)
			assert.Equal(t, ObjectId(""), s2.ID)
		}

		{
			// new driver
			var s2 Struct2
			err = bson.Unmarshal(data, &s2)
			assert.NoError(t, err)
			assert.Equal(t, ObjectId(""), s2.ID)
		}
	})

	t.Run("nil id into new non-pointer", func(t *testing.T) {
		type Struct struct {
			ID *primitive.ObjectID `bson:"_id"`
		}

		data, err := Marshal(Struct{ID: nil})
		assert.NoError(t, err)

		type Struct2 struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		// should error for new type
		{
			// old driver
			var s2 Struct2
			err = Unmarshal(data, &s2)
			assert.Error(t, err)
		}

		{
			// new driver
			var s2 Struct2
			err = bson.Unmarshal(data, &s2)
			assert.Error(t, err)
		}
	})
}

func TestMarshal_Regex(t *testing.T) {
	t.Run("pointer regex", func(t *testing.T) {
		rp := primitive.Regex{Pattern: ".*", Options: "i"}
		rb := RegEx{Pattern: ".*", Options: "i"}

		p := PStructPointer{Regex: &rp}
		b := BStructPointer{Regex: &rb}

		CheckMarshal(t, p, b)
	})

	t.Run("nil regex", func(t *testing.T) {
		p := PStructPointer{Regex: nil}
		b := BStructPointer{Regex: nil}
		CheckMarshal(t, p, b)
	})
}

func TestMarshal_time(t *testing.T) {
	now := time.Now().UTC()

	dataB, err := Marshal(bson.M{"time": now})
	if err != nil {
		t.Fatalf("marshal bson error: %v", err)
	}

	dataP, err := bson.Marshal(primitive.M{"time": now})
	if err != nil {
		t.Fatalf("marshal bson error: %v", err)
	}

	if !bytes.Equal(dataB, dataP) {
		t.Errorf("unequal bytes: %s %s", dataB, dataP)
	}

	// unmarshal
	type Struct struct {
		Time time.Time `bson:"time"`
	}

	b := Struct{}
	err = Unmarshal(dataB, &b)
	if err != nil {
		t.Fatalf("unmarshal bson error: %v", err)
	}

	p := Struct{}
	err = bson.Unmarshal(dataB, &p)
	if err != nil {
		t.Fatalf("unmarshal bson error: %v", err)
	}

	if !b.Time.UTC().Equal(p.Time.UTC()) {
		t.Errorf("unequal time: %s %s", b.Time.UTC(), p.Time.UTC())
	}
}

func TestMarshal_M(t *testing.T) {
	t.Run("basic key value", func(t *testing.T) {
		p := primitive.M{"name": "abc"}
		b := M{"name": "abc"}

		CheckMarshal(t, p, b)
	})

	t.Run("with primitive id", func(t *testing.T) {
		id := primitive.NewObjectID()

		p := primitive.M{"_id": id}
		b := M{"_id": id}

		CheckMarshal(t, p, b)
	})

	t.Run("with old id", func(t *testing.T) {
		id := NewObjectId()

		p := primitive.M{"_id": id}
		b := M{"_id": id}

		CheckMarshal(t, p, b)
	})
}

func TestMarshal_D(t *testing.T) {
	t.Run("basic key value", func(t *testing.T) {
		p := primitive.D{
			primitive.E{Key: "name", Value: "abc"},
		}

		b := D{
			DocElem{Name: "name", Value: "abc"},
		}

		CheckMarshalAndUnmarshal(t, p, b)
	})
}

func TestMarshal_A(t *testing.T) {
	t.Run("marsha ", func(t *testing.T) {
		pid := primitive.NewObjectID()
		bid := NewObjectId()

		p := primitive.M{
			"array": primitive.A{
				pid,
				bid,
				"asdf",
				primitive.M{"name": int32(123)},
				M{"name": int32(123)},
			},
		}

		b := M{
			"array": []interface{}{
				pid,
				bid,
				"asdf",
				primitive.M{"name": int32(123)},
				M{"name": int32(123)},
			},
		}

		CheckMarshal(t, p, b)
	})

	t.Run("unmarshal", func(t *testing.T) {
		pid := primitive.NewObjectID()
		bid := NewObjectId()

		p := primitive.M{
			"array": primitive.A{
				pid,
				bid,
				"asdf",
				primitive.M{"name": int32(123)},
				M{"name": int32(123)},
			},
		}

		data, err := Marshal(p)
		assert.NoError(t, err)

		t.Run("with old driver old types", func(t *testing.T) {
			m := M{}
			err = Unmarshal(data, &m)

			assert.NoError(t, err)
			expected := M{
				"array": []interface{}{
					ObjectIdHex(pid.Hex()),
					bid,
					"asdf",
					M{"name": 123},
					M{"name": 123},
				},
			}

			assert.Equal(t, expected, m)
		})

		t.Run("with old driver new types", func(t *testing.T) {
			m := primitive.M{}
			err = Unmarshal(data, &m)

			assert.NoError(t, err)

			// why do we get old A but new M?
			// The old driver has a feature that if the "base type" is a map
			// all nested maps will be the same type.
			expected := primitive.M{
				"array": []interface{}{
					ObjectIdHex(pid.Hex()),
					bid,
					"asdf",
					primitive.M{"name": 123}, // why primitive.M here, see comment above
					primitive.M{"name": 123},
				},
			}

			assert.Equal(t, expected, m)
		})

		t.Run("with new driver old types", func(t *testing.T) {
			m := M{}
			err = bson.Unmarshal(data, &m)
			assert.NoError(t, err)

			eid, err := primitive.ObjectIDFromHex(bid.Hex())
			assert.NoError(t, err)

			// why do we get old m, but new A?
			// The old driver has a feature that if the "base type" is a map
			// all nested maps will be the same type.
			expected := M{
				"array": primitive.A{
					pid,
					eid,
					"asdf",
					M{"name": int32(123)},
					M{"name": int32(123)},
				},
			}

			assert.Equal(t, expected, m)
		})

		t.Run("with new driver new types", func(t *testing.T) {
			m := primitive.M{}
			err = bson.Unmarshal(data, &m)
			assert.NoError(t, err)

			eid, err := primitive.ObjectIDFromHex(bid.Hex())
			assert.NoError(t, err)

			expected := primitive.M{
				"array": primitive.A{
					pid,
					eid,
					"asdf",
					primitive.M{"name": int32(123)},
					primitive.M{"name": int32(123)},
				},
			}

			assert.Equal(t, expected, m)
		})
	})
}

func CheckMarshalAndUnmarshal(t *testing.T, p interface{}, b interface{}) {
	checkMarshal(t, p, b, false)
}

func CheckMarshal(t *testing.T, p, b interface{}) {
	checkMarshal(t, p, b, true)
}

func checkMarshal(t *testing.T, p, b interface{}, skipUnmarshal bool) {
	// marshal bson
	bdataP, err := Marshal(p)
	assert.NoError(t, err)

	bdataB, err := Marshal(b)
	assert.NoError(t, err)

	// marshal new driver bson
	pdataP, err := bson.Marshal(p)
	assert.NoError(t, err)

	pdataB, err := bson.Marshal(b)
	assert.NoError(t, err)

	checkRaw(t, bdataP, bdataB)
	checkRaw(t, pdataP, pdataB)
	checkRaw(t, pdataB, pdataP)

	t.Logf("p with b %v", bdataP)
	t.Logf("b with b %v", bdataB)
	t.Logf("p with p %v", pdataP)
	t.Logf("b with p %v", pdataB)

	// unmarshal
	if skipUnmarshal {
		return
	}
	data := pdataP

	t.Run("unmarshal into primitive with old diver", func(t *testing.T) {
		checkUnmarshal(t, Unmarshal, data, pointerTo(p), newPointerTo(p))
	})

	t.Run("unmarshal into primitive with new diver", func(t *testing.T) {
		checkUnmarshal(t, bson.Unmarshal, data, pointerTo(p), newPointerTo(p))
	})

	t.Run("unmarshal into bson with old diver", func(t *testing.T) {
		checkUnmarshal(t, Unmarshal, data, pointerTo(b), newPointerTo(b))
	})

	t.Run("unmarshal into bson with new diver", func(t *testing.T) {
		checkUnmarshal(t, bson.Unmarshal, data, pointerTo(b), newPointerTo(b))
	})
}

func checkUnmarshal(
	t *testing.T,
	unmarshaler func([]byte, interface{}) (err error),
	data []byte, existing, newP interface{},
) {
	assert.NotEqual(t, fmt.Sprintf("%p", existing), fmt.Sprintf("%p", newP))

	err := unmarshaler(data, newP)
	assert.NoError(t, err)

	assert.Equal(t, existing, newP)
}

func checkRaw(t *testing.T, d1, d2 []byte) {
	t.Helper()
	if !bytes.Equal(d1, d2) {
		t.Logf("%v", d1)
		t.Logf("%v", d2)
		t.Errorf("data not marshed to the same bytes")
	}
}

func pointerTo(v interface{}) interface{} {
	p := reflect.New(reflect.TypeOf(v))
	p.Elem().Set(reflect.ValueOf(v))

	return p.Interface()
}

func newPointerTo(v interface{}) interface{} {
	val := reflect.ValueOf(v)
	p := reflect.New(val.Type()).Interface()

	return reflect.ValueOf(p).Interface()
}
