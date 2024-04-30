package unpack

import (
	"github.com/almerlucke/muse"
	"reflect"
)

type Unpack struct {
	*muse.BaseControl
}

func New() *Unpack {
	u := &Unpack{
		BaseControl: muse.NewBaseControl(),
	}

	u.SetSelf(u)

	return u
}

func (u *Unpack) ReceiveControlValue(value any, index int) {
	if index == 0 {
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice:
			fallthrough
		case reflect.Array:
			l := rv.Len()
			for i := range l {
				ri := l - i - 1
				u.SendControlValue(rv.Index(ri).Interface(), ri)
			}
		default:
			break
		}
	}
}

func (u *Unpack) ReceiveMessage(msg any) []*muse.Message {
	rv := reflect.ValueOf(msg)
	switch rv.Kind() {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		l := rv.Len()
		for i := range l {
			ri := l - i - 1
			u.SendControlValue(rv.Index(ri), ri)
		}
	default:
		break
	}

	return nil
}
