// Package grate opens tabular data files (such as spreadsheets and delimited plaintext files)
// and allows programmatic access to the data contents in a consistent interface.
package grate

import (
	"errors"
	"log"
	"sort"
)

// Source represents a set of data collections.
type Source interface {
	// List the individual data tables within this source.
	List() ([]string, error)

	// Get a Collection from the source by name.
	Get(name string) (Collection, error)

	// Close the source and discard memory.
	Close() error
}

// Collection represents an iterable collection of records.
type Collection interface {
	// Next advances to the next record of content.
	// It MUST be called prior to any Scan().
	Next() bool

	// Strings extracts values from the current record into a list of strings.
	Strings() []string

	// Types extracts the data types from the current record into a list.
	// options: "boolean", "integer", "float", "string", "date",
	// and special cases: "blank", "hyperlink" which are string types
	Types() []string

	// Formats extracts the format codes for the current record into a list.
	Formats() []string

	// Scan extracts values from the current record into the provided arguments
	// Arguments must be pointers to one of 5 supported types:
	//     bool, int64, float64, string, or time.Time
	// If invalid, returns ErrInvalidScanType
	Scan(args ...interface{}) error

	// IsEmpty returns true if there are no data values.
	IsEmpty() bool

	// Err returns the last error that occured.
	Err() error
}

// OpenFunc defines a Source's instantiation function.
// It should return ErrNotInFormat immediately if filename is not of the correct file type.
type OpenFunc func(filename string) (Source, error)

// Open a tabular data file and return a Source for accessing it's contents.
func Open(filename string) (Source, error) {
	for _, o := range srcTable {
		src, err := o.op(filename)
		if err == nil {
			return src, nil
		}
		if !errors.Is(err, ErrNotInFormat) {
			return nil, err
		}
		if Debug {
			log.Println(" ", filename, "is not in", o.name, "format")
		}
	}
	return nil, ErrUnknownFormat
}

type srcOpenTab struct {
	name string
	pri  int
	op   OpenFunc
}

var srcTable = make([]*srcOpenTab, 0, 20)

// Register the named source as a grate datasource implementation.
func Register(name string, priority int, opener OpenFunc) error {
	if Debug {
		log.Println("Registering the", name, "format at priority", priority)
	}
	srcTable = append(srcTable, &srcOpenTab{name: name, pri: priority, op: opener})
	sort.Slice(srcTable, func(i, j int) bool {
		return srcTable[i].pri < srcTable[j].pri
	})
	return nil
}

const (
	// ContinueColumnMerged marks a continuation column within a merged cell.
	ContinueColumnMerged = "→"
	// EndColumnMerged marks the last column of a merged cell.
	EndColumnMerged = "⇥"

	// ContinueRowMerged marks a continuation row within a merged cell.
	ContinueRowMerged = "↓"
	// EndRowMerged marks the last row of a merged cell.
	EndRowMerged = "⤓"
)
