package sq

import (
	"github.com/jmoiron/sqlx"
	"time"
)
type ScanFunc func(rows *sqlx.Rows) error

type UintLister interface {
	Append(i uint)
}
func ScanUintLister(list UintLister) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item uint
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		list.Append(item)
		return nil
	}
}
type IntLister interface {
	Append(i int)
}
func ScanIntLister(list IntLister) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item int
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		list.Append(item)
		return nil
	}
}
type BytesIDLister interface {
	Append(data []byte)
}
func ScanBytesLister(list BytesIDLister) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item []byte
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		list.Append(item)
		return nil
	}
}
type StringLister interface {
	Append(s string)
}
func ScanStringLister(list StringLister) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item string
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		list.Append(item)
		return nil
	}
}
func ScanBytes(bytes *[][]byte) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item []byte
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		*bytes = append(*bytes, item)
		return nil
	}
}
func ScanStrings(strings *[]string) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item string
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		*strings = append(*strings, item)
		return nil
	}
}
func ScanInts(ints *[]int) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item int
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		*ints = append(*ints, item)
		return nil
	}
}
func ScanBool(bools *[]bool) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item bool
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		*bools = append(*bools, item)
		return nil
	}
}
func ScanTimes(times *[]time.Time) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item time.Time
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		*times = append(*times, item)
		return nil
	}
}
func ScanUints(uints *[]uint) ScanFunc {
	return func(rows *sqlx.Rows) error {
		var item uint
		err := rows.Scan(&item) ; if err != nil {
			return err
		}
		*uints = append(*uints, item)
		return nil
	}
}
