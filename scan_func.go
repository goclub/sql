package sq

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type Scaner func(rows *sqlx.Rows) error

type UintLister interface {
	Append(i uint)
}

func ScanUintLister(list UintLister) Scaner {
	return func(rows *sqlx.Rows) error {
		var item uint
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		list.Append(item)
		return nil
	}
}

type IntLister interface {
	Append(i int)
}

func ScanIntLister(list IntLister) Scaner {
	return func(rows *sqlx.Rows) error {
		var item int
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		list.Append(item)
		return nil
	}
}

type BytesIDLister interface {
	Append(data []byte)
}

func ScanBytesLister(list BytesIDLister) Scaner {
	return func(rows *sqlx.Rows) error {
		var item []byte
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		list.Append(item)
		return nil
	}
}

type StringLister interface {
	Append(s string)
}

func ScanStringLister(list StringLister) Scaner {
	return func(rows *sqlx.Rows) error {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		list.Append(item)
		return nil
	}
}
func ScanBytes(bytes *[][]byte) Scaner {
	return func(rows *sqlx.Rows) error {
		var item []byte
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		*bytes = append(*bytes, item)
		return nil
	}
}
func ScanStrings(strings *[]string) Scaner {
	return func(rows *sqlx.Rows) error {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		*strings = append(*strings, item)
		return nil
	}
}
func ScanInts(ints *[]int) Scaner {
	return func(rows *sqlx.Rows) error {
		var item int
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		*ints = append(*ints, item)
		return nil
	}
}
func ScanUints(ints *[]uint) Scaner {
	return func(rows *sqlx.Rows) error {
		var item uint
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		*ints = append(*ints, item)
		return nil
	}
}
func ScanInt64s(ints *[]int64) Scaner {
	return func(rows *sqlx.Rows) error {
		var item int64
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		*ints = append(*ints, item)
		return nil
	}
}
func ScanUint64s(ints *[]uint64) Scaner {
	return func(rows *sqlx.Rows) error {
		var item uint64
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		*ints = append(*ints, item)
		return nil
	}
}
func ScanBool(bools *[]bool) Scaner {
	return func(rows *sqlx.Rows) error {
		var item bool
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		*bools = append(*bools, item)
		return nil
	}
}
func ScanTimes(times *[]time.Time) Scaner {
	return func(rows *sqlx.Rows) error {
		var item time.Time
		err := rows.Scan(&item)
		if err != nil {
			return err
		}
		*times = append(*times, item)
		return nil
	}
}
