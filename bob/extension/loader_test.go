package extension

import (
	"context"
	"testing"

	"github.com/tetratelabs/wazero"
)

// minimalWASM is a hand-encoded WASM module that:
//   - has 1 page of linear memory exported as "memory"
//   - exports _init() i32, _on_event(i32,i32) i32, _alloc(i32) i32, _free(i32)
//   - all functions return 0 / do nothing
//   - has NO imports
var minimalWASM = []byte{
	// magic + version
	0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00,
	// type section: 4 types
	0x01, 0x14, // section id=1, size=20
	0x04,                               // count=4
	0x60, 0x00, 0x01, 0x7F,             // type 0: () -> i32
	0x60, 0x02, 0x7F, 0x7F, 0x01, 0x7F, // type 1: (i32, i32) -> i32
	0x60, 0x01, 0x7F, 0x01, 0x7F, // type 2: (i32) -> i32
	0x60, 0x01, 0x7F, 0x00, // type 3: (i32) -> ()
	// function section: 4 functions
	0x03, 0x05, // section id=3, size=5
	0x04,                   // count=4
	0x00, 0x01, 0x02, 0x03, // type indices
	// memory section: 1 page
	0x05, 0x03, // section id=5, size=3
	0x01,       // count=1
	0x00, 0x01, // min=1, no max
	// export section: 5 exports
	0x07, 0x2F, // section id=7, size=47
	0x05, // count=5
	// "memory", memory, 0
	0x06, 0x6D, 0x65, 0x6D, 0x6F, 0x72, 0x79, 0x02, 0x00,
	// "_init", func, 0
	0x05, 0x5F, 0x69, 0x6E, 0x69, 0x74, 0x00, 0x00,
	// "_on_event", func, 1
	0x09, 0x5F, 0x6F, 0x6E, 0x5F, 0x65, 0x76, 0x65, 0x6E, 0x74, 0x00, 0x01,
	// "_alloc", func, 2
	0x06, 0x5F, 0x61, 0x6C, 0x6C, 0x6F, 0x63, 0x00, 0x02,
	// "_free", func, 3
	0x05, 0x5F, 0x66, 0x72, 0x65, 0x65, 0x00, 0x03,
	// code section: 4 bodies
	0x0A, 0x13, // section id=10, size=19
	0x04,                               // count=4
	0x04, 0x00, 0x41, 0x00, 0x0B, // body 0 (_init): i32.const 0, end
	0x04, 0x00, 0x41, 0x00, 0x0B, // body 1 (_on_event): i32.const 0, end
	0x04, 0x00, 0x41, 0x00, 0x0B, // body 2 (_alloc): i32.const 0, end
	0x02, 0x00, 0x0B, // body 3 (_free): end
}

// missingFreeWASM is like minimalWASM but exports only _init, _on_event, _alloc (no _free).
var missingFreeWASM = []byte{
	// magic + version
	0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00,
	// type section: 3 types, size=16
	// count(1) + type0(4) + type1(6) + type2(5) = 16
	0x01, 0x10,
	0x03,                               // count=3
	0x60, 0x00, 0x01, 0x7F,             // type 0: () -> i32
	0x60, 0x02, 0x7F, 0x7F, 0x01, 0x7F, // type 1: (i32, i32) -> i32
	0x60, 0x01, 0x7F, 0x01, 0x7F, // type 2: (i32) -> i32
	// function section: 3 functions, size=4
	0x03, 0x04,
	0x03,             // count=3
	0x00, 0x01, 0x02, // type indices
	// memory section: size=3
	0x05, 0x03, 0x01, 0x00, 0x01,
	// export section: 4 exports (no _free), size=39
	// count(1) + memory(9) + _init(8) + _on_event(12) + _alloc(9) = 39
	0x07, 0x27,
	0x04, // count=4
	// "memory" (6), memory=2, index=0
	0x06, 0x6D, 0x65, 0x6D, 0x6F, 0x72, 0x79, 0x02, 0x00,
	// "_init" (5), func=0, index=0
	0x05, 0x5F, 0x69, 0x6E, 0x69, 0x74, 0x00, 0x00,
	// "_on_event" (9), func=0, index=1
	0x09, 0x5F, 0x6F, 0x6E, 0x5F, 0x65, 0x76, 0x65, 0x6E, 0x74, 0x00, 0x01,
	// "_alloc" (6), func=0, index=2
	0x06, 0x5F, 0x61, 0x6C, 0x6C, 0x6F, 0x63, 0x00, 0x02,
	// code section: 3 bodies, size=16
	// count(1) + body0(5) + body1(5) + body2(5) = 16
	0x0A, 0x10,
	0x03,
	0x04, 0x00, 0x41, 0x00, 0x0B, // body 0: i32.const 0, end
	0x04, 0x00, 0x41, 0x00, 0x0B, // body 1
	0x04, 0x00, 0x41, 0x00, 0x0B, // body 2
}

func TestValidateExports_AllPresent(t *testing.T) {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	mod, err := r.InstantiateWithConfig(ctx, minimalWASM,
		wazero.NewModuleConfig().WithName("test-validate").WithStartFunctions())
	if err != nil {
		t.Fatalf("instantiate: %v", err)
	}
	defer mod.Close(ctx)

	if err := validateExports(mod); err != nil {
		t.Errorf("validateExports returned unexpected error: %v", err)
	}
}

func TestValidateExports_MissingFree(t *testing.T) {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	mod, err := r.InstantiateWithConfig(ctx, missingFreeWASM,
		wazero.NewModuleConfig().WithName("test-missing-free").WithStartFunctions())
	if err != nil {
		t.Fatalf("instantiate: %v", err)
	}
	defer mod.Close(ctx)

	err = validateExports(mod)
	if err == nil {
		t.Fatal("expected error for missing _free export, got nil")
	}
}

func TestCallInit_ReturnsZero(t *testing.T) {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	mod, err := r.InstantiateWithConfig(ctx, minimalWASM,
		wazero.NewModuleConfig().WithName("test-callinit").WithStartFunctions())
	if err != nil {
		t.Fatalf("instantiate: %v", err)
	}
	defer mod.Close(ctx)

	if err := callInit(ctx, mod); err != nil {
		t.Errorf("callInit returned unexpected error: %v", err)
	}
}
