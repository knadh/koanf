package koanf_test

import (
	"testing"

	"github.com/knadh/koanf/maps"
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

func TestFlatten(t *testing.T) {
	f, k := maps.Flatten(testMap, nil, delim)
	assert.Equal(t, map[string]interface{}{
		"parent.child.key":            123,
		"parent.child.key~1with~1dot": 456,
		"top":                         789,
		"empty":                       map[string]interface{}{},
	}, f)
	assert.Equal(t, map[string][]string{
		"parent.child.key":            {"parent", "child", "key"},
		"parent.child.key~1with~1dot": {"parent", "child", "key.with.dot"},
		"top":                         {"top"},
		"empty":                       {"empty"},
	}, k)
}

func BenchmarkFlatten(b *testing.B) {
	for n := 0; n < b.N; n++ {
		maps.Flatten(testMap3, nil, delim)
	}
}

func TestUnflatten(t *testing.T) {
	m, _ := maps.Flatten(testMap, nil, delim)
	um := maps.Unflatten(m, delim)
	assert.Equal(t, testMap, um)

	m, _ = maps.Flatten(testMap2, nil, delim)
	um = maps.Unflatten(m, delim)
	assert.Equal(t, testMap2, um)
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
	maps.IntfaceKeysToStrings(m)
	assert.Equal(t, testMap2, m)
}

func TestMapMerge(t *testing.T) {
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
	maps.Merge(m2, m1)

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

func TestMapMerge2(t *testing.T) {
	src := map[string]interface{}{
		"globals": map[string]interface{}{
			"features": map[string]interface{}{
				"testing": map[string]interface{}{
					"enabled": false,
				},
			},
		},
	}

	dest := map[string]interface{}{
		"globals": map[string]interface{}{
			"features": map[string]interface{}{
				"testing": map[string]interface{}{
					"enabled":    true,
					"anotherKey": "value",
				},
			},
		},
	}

	maps.Merge(src, dest)
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
	err := maps.MergeStrict(m2, m1)
	assert.Error(t, err)
}

func TestMapDelete(t *testing.T) {
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

	maps.Delete(testMap, []string{"parent", "child"})
	assert.Equal(t, map[string]interface{}{
		"top":   789,
		"empty": map[string]interface{}{},
	}, testMap)

	maps.Delete(testMap2, []string{"list"})
	maps.Delete(testMap2, []string{"empty"})
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
	assert.Equal(t, 123, maps.Search(testMap, []string{"parent", "child", "key"}))
	assert.Equal(t, map[string]interface{}{
		"key":          123,
		"key.with.dot": 456,
	}, maps.Search(testMap, []string{"parent", "child"}))
	assert.Equal(t, 456, maps.Search(testMap, []string{"parent", "child", "key.with.dot"}))
	assert.Equal(t, 789, maps.Search(testMap, []string{"top"}))
	assert.Equal(t, map[string]interface{}{}, maps.Search(testMap, []string{"empty"}))
	assert.Nil(t, maps.Search(testMap, []string{"xxx", "xxx"}))
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
	assert.Equal(t, mp, maps.Copy(mp))
}

func TestLookupMaps(t *testing.T) {
	assert.Equal(t, map[string]bool{"a": true, "b": true}, maps.StringSliceToLookupMap([]string{"a", "b"}))
	assert.Equal(t, map[string]bool{}, maps.StringSliceToLookupMap(nil))
	assert.Equal(t, map[int64]bool{1: true, 2: true}, maps.Int64SliceToLookupMap([]int64{1, 2}))
	assert.Equal(t, map[int64]bool{}, maps.Int64SliceToLookupMap(nil))

}
