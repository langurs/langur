// langur/vm/process/builtins.go

package process

import (
	"langur/object"
)

func GetBuiltInByName(name string) *object.BuiltIn {
	for _, bi := range BuiltIns {
		if bi.FnSignature.Name == name {
			return bi
		}
	}
	return nil
}

func GetBuiltInImpurityStatus(name string) bool {
	for _, bi := range BuiltIns {
		if bi.FnSignature.Name == name {
			return bi.HasImpureEffects()
		}
	}
	return false
}

type BuiltInFunction = func(pr *Process, args ...object.Object) object.Object

// the index of built-ins
var BuiltIns = []*object.BuiltIn{
	// internal built-ins
	bi__limit,
	bi__values,
	bi__keys,
	bi__len,
	bi__ishash,

	// type conversion functions
	bi_string,
	bi_number,
	bi_complex,
	bi_hash,
	bi_datetime,
	bi_duration,
	bi_bool,

	// external built-ins
	bi_abs,
	bi_ceiling,
	bi_floor,
	bi_gcd,
	bi_lcm,
	bi_max,
	bi_min,
	bi_minmax,
	bi_mean,
	bi_mid,
	bi_round,
	bi_trunc,
	bi_simplify,

	bi_all,
	bi_any,

	bi_execT,
	bi_execTH,

	bi_exit,

	bi_count,
	bi_filter,

	bi_fold,
	bi_zip,

	bi_keys,

	bi_join,

	bi_lcase,
	bi_tcase,
	bi_ucase,

	bi_len,

	bi_less,
	bi_more,
	bi_reverse,
	bi_rotate,

	bi_trim,
	bi_ltrim,
	bi_rtrim,

	bi_map,
	bi_mapX,

	bi_tran,
	bi_replace,
	bi_matching,
	bi_match,
	bi_matches,
	bi_submatch,
	bi_submatchH,
	bi_submatches,
	bi_submatchesH,

	bi_nfc,
	bi_nfd,
	bi_nfkc,
	bi_nfkd,

	bi_nn,

	bi_cd,
	bi_prop,

	bi_random,

	bi_reCompile,
	bi_reEsc,

	bi_s2b,
	bi_b2s,
	bi_cp2s,
	bi_s2cp,
	bi_s2gc,
	bi_s2s,
	bi_s2n,

	bi_series,

	bi_sin,
	bi_cos,
	bi_tan,
	bi_atan,

	bi_sleep,
	bi_ticks,

	bi_sort,

	bi_split,
	bi_index,
	bi_indices,
	bi_subindex,
	bi_subindices,

	bi_read,
	bi_write,
	bi_writeln,
	bi_writeErr,
	bi_writelnErr,

	bi_readfile,
	bi_writefile,
	bi_appendfile,
}
