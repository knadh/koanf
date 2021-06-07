package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testMap = map[string]interface{}{
	"parent": map[string]interface{}{
		"child": map[string]interface{}{
			"key":          123,
			"key.with.dot": 456,
		},
	},
	"top":   789,
	"empty": map[string]interface{}{},
}
var testMap2 = map[string]interface{}{
	"list": []interface{}{
		map[string]interface{}{
			"child": map[string]interface{}{
				"key": 123,
			},
		},
		map[string]interface{}{
			"child": map[string]interface{}{
				"key": 123,
			},
		},
	},
	"parent": map[string]interface{}{
		"child": map[string]interface{}{
			"key": 123,
		},
	},
	"top":   789,
	"empty": map[string]interface{}{},
}

var testMap3 = map[string]interface{}{
	"list": []interface{}{
		map[string]interface{}{
			"child": map[string]interface{}{
				"key": 123,
				"child": map[string]interface{}{
					"key": 123,
					"child": map[string]interface{}{
						"key": 123,
						"child": map[string]interface{}{
							"key": 123,
							"child": map[string]interface{}{
								"key": 123,
								"child": map[string]interface{}{
									"key": 123,
									"child": map[string]interface{}{
										"key": 123,
									},
								},
							},
						},
					},
				},
			},
		},
		map[string]interface{}{
			"child": map[string]interface{}{
				"key": 123,
				"child": map[string]interface{}{
					"key": 123,
				},
			},
		},
	},
	"parent": map[string]interface{}{
		"child": map[string]interface{}{
			"key": 123,
			"child": map[string]interface{}{
				"key": 123,
				"child": map[string]interface{}{
					"key": 123,
					"child": map[string]interface{}{
						"key": 123,
						"child": map[string]interface{}{
							"key": 123,
						},
					},
				},
			},
		},
	},
	"top": 789,
	"child": map[string]interface{}{
		"key": 123,
		"child": map[string]interface{}{
			"key": 123,
		},
	},
	"empty": map[string]interface{}{},
}

const delim = "."

func TestFlatten(t *testing.T) {
	f, k := Flatten(testMap, nil, delim)
	assert.Equal(t, map[string]interface{}{
		"parent.child.key":          123,
		"parent.child.key.with.dot": 456,
		"top":                       789,
		"empty":                     map[string]interface{}{},
	}, f)
	assert.Equal(t, map[string][]string{
		"parent.child.key":          {"parent", "child", "key"},
		"parent.child.key.with.dot": {"parent", "child", "key.with.dot"},
		"top":                       {"top"},
		"empty":                     {"empty"},
	}, k)
}

func BenchmarkFlatten(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Flatten(testMap3, nil, delim)
	}
}

func TestUnflatten(t *testing.T) {
	m, _ := Flatten(testMap, nil, delim)
	um := Unflatten(m, delim)
	assert.NotEqual(t, um, testMap)

	m, _ = Flatten(testMap2, nil, delim)
	um = Unflatten(m, delim)
	assert.Equal(t, um, testMap2)
}

func TestIntfaceKeysToStrings(t *testing.T) {
	m := map[string]interface{}{
		"list": []interface{}{
			map[interface{}]interface{}{
				"child": map[interface{}]interface{}{
					"key": 123,
				},
			},
			map[interface{}]interface{}{
				"child": map[interface{}]interface{}{
					"key": 123,
				},
			},
		},
		"parent": map[interface{}]interface{}{
			"child": map[interface{}]interface{}{
				"key": 123,
			},
		},
		"top":   789,
		"empty": map[interface{}]interface{}{},
	}
	IntfaceKeysToStrings(m)
	assert.Equal(t, testMap2, m)
}

func TestMerge(t *testing.T) {
	m1 := map[string]interface{}{
		"parent": map[string]interface{}{
			"child": map[string]interface{}{
				"key": 123,
			},
			"child2": map[string]interface{}{
				"key": 123,
			},
		},
		"top":   789,
		"empty": map[string]interface{}{},
		"key":   1,
	}
	m2 := map[string]interface{}{
		"parent": map[string]interface{}{
			"child": map[string]interface{}{
				"key": 456,
				"val": 789,
			},
		},
		"child": map[string]interface{}{
			"key": 456,
		},
		"newtop": 999,
		"empty":  []int{1, 2, 3},
		"key":    "string",
	}
	Merge(m2, m1)

	out := map[string]interface{}{
		"parent": map[string]interface{}{
			"child": map[string]interface{}{
				"key": 456,
				"val": 789,
			},
			"child2": map[string]interface{}{
				"key": 123,
			},
		},
		"child": map[string]interface{}{
			"key": 456,
		},
		"top":    789,
		"newtop": 999,
		"empty":  []int{1, 2, 3},
		"key":    "string",
	}
	assert.Equal(t, out, m1)
}

func TestMergeStrict(t *testing.T) {
	m1 := map[string]interface{}{
		"parent": map[string]interface{}{
			"child": map[string]interface{}{
				"key": "123",
			},
			"child2": map[string]interface{}{
				"key": 123,
			},
		},
		"top":   789,
		"empty": []int{},
		"key":   1,
	}
	m2 := map[string]interface{}{
		"parent": map[string]interface{}{
			"child": map[string]interface{}{
				"key": 456,
				"val": 789,
			},
		},
		"child": map[string]interface{}{
			"key": 456,
		},
		"newtop": 999,
		"empty":  []int{1, 2, 3},
		"key":    "string",
	}
	err := MergeStrict(m2, m1)
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	testMap := map[string]interface{}{
		"parent": map[string]interface{}{
			"child": map[string]interface{}{
				"key":          123,
				"key.with.dot": 456,
			},
		},
		"top":   789,
		"empty": map[string]interface{}{},
	}
	testMap2 := map[string]interface{}{
		"list": []interface{}{
			map[string]interface{}{
				"child": map[string]interface{}{
					"key": 123,
				},
			},
			map[string]interface{}{
				"child": map[string]interface{}{
					"key": 123,
				},
			},
		},
		"parent": map[string]interface{}{
			"child": map[string]interface{}{
				"key": 123,
			},
		},
		"top":   789,
		"empty": map[string]interface{}{},
	}

	Delete(testMap, []string{"parent", "child"})
	assert.Equal(t, map[string]interface{}{
		"top":   789,
		"empty": map[string]interface{}{},
	}, testMap)

	Delete(testMap2, []string{"list"})
	Delete(testMap2, []string{"empty"})
	assert.Equal(t, map[string]interface{}{
		"parent": map[string]interface{}{
			"child": map[string]interface{}{
				"key": 123,
			},
		},
		"top": 789,
	}, testMap2)
}

func TestSearch(t *testing.T) {
	assert.Equal(t, 123, Search(testMap, []string{"parent", "child", "key"}))
	assert.Equal(t, map[string]interface{}{
		"key":          123,
		"key.with.dot": 456,
	}, Search(testMap, []string{"parent", "child"}))
	assert.Equal(t, 456, Search(testMap, []string{"parent", "child", "key.with.dot"}))
	assert.Equal(t, 789, Search(testMap, []string{"top"}))
	assert.Equal(t, map[string]interface{}{}, Search(testMap, []string{"empty"}))
	assert.Nil(t, Search(testMap, []string{"xxx", "xxx"}))
}

func TestCopy(t *testing.T) {
	mp := map[string]interface{}{
		"parent": map[string]interface{}{
			"child": map[string]interface{}{
				"key":          float64(123),
				"key.with.dot": float64(456),
			},
		},
		"top":   float64(789),
		"empty": map[string]interface{}{},
	}
	assert.Equal(t, mp, Copy(mp))
}

func TestLookupMaps(t *testing.T) {
	assert.Equal(t, map[string]bool{"a": true, "b": true}, StringSliceToLookupMap([]string{"a", "b"}))
	assert.Equal(t, map[string]bool{}, StringSliceToLookupMap(nil))
	assert.Equal(t, map[int64]bool{1: true, 2: true}, Int64SliceToLookupMap([]int64{1, 2}))
	assert.Equal(t, map[int64]bool{}, Int64SliceToLookupMap(nil))

}
