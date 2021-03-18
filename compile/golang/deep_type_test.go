package golang

import (
	"fmt"
	"testing"

	"github.com/fioncat/go-gendb/misc/term"
)

func runTypeTests(ts []string) {
	for _, t := range ts {
		dt, err := ParseDeepType(t)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Type: %s >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n", t)
		term.Show(dt)
		fmt.Println()
	}
}

func TestDeepTypeSimple(t *testing.T) {
	tests := []string{
		"int64", "string", "uint32",
		"uint64", "float32", "float64",
	}
	runTypeTests(tests)
}

func TestDeepTypeSimplePointer(t *testing.T) {
	tests := []string{
		"*int64", "*string", "*uint32",
		"*uint64", "*float32", "*float64",
	}
	runTypeTests(tests)
}

func TestDeepTypeStruct(t *testing.T) {
	tests := []string{
		"User",
		"*User",
		"Detail",
		"*Detail",
		"model.Base",
		"*model.Base",
		"golang.Interface",
		"*parser.Result",
		"coder.Coder",
		"bson.ObjectId",
	}
	runTypeTests(tests)
}

func TestDeepTypeSliceSimple0(t *testing.T) {
	tests := []string{
		"[]string", "[]int32", "[]int", "[]int64",
		"[]float32", "[]float64",
	}
	runTypeTests(tests)
}

func TestDeepTypeSliceSimple1(t *testing.T) {
	tests := []string{
		"[]User",
		"[]*User",
		"[]Detail",
		"[]*Detail",
		"[]model.Base",
		"[]*model.Base",
		"[]golang.Interface",
		"[]*parser.Result",
		"[]coder.Coder",
	}
	runTypeTests(tests)
}

func TestDeepTypeSliceComplex0(t *testing.T) {
	tests := []string{
		"[][]User",
		"[][]*User",
		"[][][]Detail",
		"[][][]*Detail",
		"[]map[string]*Detail",
		"[][]map[int64]golang.Interface",
		"[]map[int64][]*model.Prop",
		"[][]int32",
		"[][]string",
		"[][][]float64",
	}
	runTypeTests(tests)
}

func TestDeepTypeMap(t *testing.T) {
	tests := []string{
		"map[string]string",
		"map[string]interface{}",
		"map[int32]string",
		"map[int64]*model.User",
		"map[string]*Detail",
		"map[int]model.User",
		"map[uint32]yml.Object",
	}
	runTypeTests(tests)
}

func TestDeepTypeMapComplex(t *testing.T) {
	tests := []string{
		"map[string][]*Detail",
		"map[string][]model.Detail",
		"map[int32][]*model.Detail",
		"map[int64]map[string][][]model.Detail",
		"map[string]map[int32]string",
		"map[int][]string",
		"map[uint32][]yml.Object",
	}
	runTypeTests(tests)
}
