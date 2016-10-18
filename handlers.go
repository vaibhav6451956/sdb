package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/francescomari/sdb/binaries"
	"github.com/francescomari/sdb/graph"
	"github.com/francescomari/sdb/index"
	"github.com/francescomari/sdb/segment"
)

var errInvalidFormat = errors.New("Invalid format")

func invalidFormat() handler {
	return func(_ string, _ io.Reader) error {
		return errInvalidFormat
	}
}

func doPrintBinaries(f format, w io.Writer) handler {
	switch f {
	case formatHex:
		return doPrintHexTo(w)
	case formatText:
		return doPrintBinariesTo(w)
	default:
		return invalidFormat()
	}
}

func doPrintBinariesTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var bns binaries.Binaries
		if _, err := bns.ReadFrom(r); err != nil {
			return err
		}
		for _, generation := range bns.Generations {
			fmt.Fprintf(w, "%d\n", generation.Generation)
			for _, segment := range generation.Segments {
				fmt.Fprintf(w, "    %s\n", segmentID(segment.Msb, segment.Lsb))
				for _, reference := range segment.References {
					fmt.Fprintf(w, "        %s\n", reference)
				}
			}
		}
		return nil
	}
}

func doPrintGraph(f format, w io.Writer) handler {
	switch f {
	case formatHex:
		return doPrintHexTo(w)
	case formatText:
		return doPrintGraphTo(w)
	default:
		return invalidFormat()
	}
}

func doPrintGraphTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var gph graph.Graph
		if _, err := gph.ReadFrom(r); err != nil {
			return nil
		}
		for _, entry := range gph.Entries {
			fmt.Fprintf(w, "%s\n", segmentID(entry.Msb, entry.Lsb))
			for _, reference := range entry.References {
				fmt.Fprintf(w, "    %s\n", segmentID(reference.Msb, reference.Lsb))
			}
		}
		return nil
	}
}

func doPrintIndex(f format, w io.Writer) handler {
	switch f {
	case formatHex:
		return doPrintHexTo(w)
	case formatText:
		return doPrintIndexTo(w)
	default:
		return invalidFormat()
	}
}

func doPrintIndexTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var idx index.Index
		if _, err := idx.ReadFrom(r); err != nil {
			return err
		}
		for _, e := range idx.Entries {
			id := segmentID(e.Msb, e.Lsb)
			fmt.Fprintf(w, "%s %s %8x %6d %6d\n", segmentType(id), id, e.Position, e.Size, e.Generation)
		}
		return nil
	}
}

func doPrintSegmentNameTo(w io.Writer) handler {
	return func(n string, _ io.Reader) error {
		id := normalizeSegmentID(entryNameToSegmentID(n))
		fmt.Fprintf(w, "%s %s\n", segmentType(id), id)
		return nil
	}
}

func doPrintSegment(f format, w io.Writer) handler {
	switch f {
	case formatHex:
		return doPrintHexTo(w)
	case formatText:
		return doPrintSegmentTo(w)
	default:
		return invalidFormat()
	}
}

func doPrintSegmentTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var s segment.Segment
		if _, err := s.ReadFrom(r); err != nil {
			return err
		}
		fmt.Fprintf(w, "Version    %d\n", s.Version)
		fmt.Fprintf(w, "Generation %d\n", s.Generation)
		fmt.Fprintf(w, "References\n")
		for i, r := range s.References {
			fmt.Fprintf(w, "    %4d %s\n", i+1, segmentID(r.Msb, r.Lsb))
		}
		fmt.Fprintf(w, "Records\n")
		for _, r := range s.Records {
			fmt.Fprintf(w, "    %08x %-10s %08x\n", r.Number, recordType(r.Type), r.Offset)
		}
		return nil
	}
}

func doPrintNameTo(w io.Writer) handler {
	return func(n string, _ io.Reader) error {
		fmt.Fprintln(w, n)
		return nil
	}
}

func doPrintHexTo(w io.Writer) handler {
	return func(_ string, r io.Reader) (err error) {
		d := hex.Dumper(w)
		defer d.Close()
		_, err = io.Copy(d, r)
		return
	}
}

func isBulkSegmentID(id string) bool {
	return id[16] == 'b'
}

func normalizeSegmentID(id string) string {
	return strings.ToLower(strings.TrimSpace(strings.Replace(id, "-", "", -1)))
}

func recordType(t segment.RecordType) string {
	switch t {
	case segment.RecordTypeBlock:
		return "block"
	case segment.RecordTypeList:
		return "list"
	case segment.RecordTypeListBucket:
		return "bucket"
	case segment.RecordTypeMapBranch:
		return "branch"
	case segment.RecordTypeMapLeaf:
		return "leaf"
	case segment.RecordTypeNode:
		return "node"
	case segment.RecordTypeTemplate:
		return "template"
	case segment.RecordTypeValue:
		return "value"
	default:
		return "unknown"
	}
}

func segmentType(id string) string {
	if isBulkSegmentID(id) {
		return "bulk"
	}
	return "data"
}

func segmentID(msb, lsb uint64) string {
	return fmt.Sprintf("%016x%016x", msb, lsb)
}
