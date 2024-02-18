package metricsql

import (
	"testing"
)

func TestPrettifyError(t *testing.T) {
	f := func(s string) {
		t.Helper()

		result, err := Prettify(s)
		if err == nil {
			t.Fatalf("expecting non-nil error")
		}
		if result != "" {
			t.Fatalf("expecting empty result; got %q", result)
		}
	}

	f(`foo{`)
	f(`invalid query`)
}

func TestPrettifySuccess(t *testing.T) {
	another := func(s, resultExpected string) {
		t.Helper()

		result, err := Prettify(s)
		if err != nil {
			t.Fatalf("unexpected error when parsing %q: %s", s, err)
		}
		if result != resultExpected {
			t.Fatalf("unexpected query after prettifying;\ngot\n%s\nwant\n%s", result, resultExpected)
		}

		// Verify that prettifying of the result gives the same result
		result2, err := Prettify(result)
		if err != nil {
			t.Fatalf("unexpected error when parsing prettified %q: %s", s, err)
		}
		if result2 != result {
			t.Fatalf("unexpected result after prettifying already prettified result;\ngot\n%s\nwant\n%s", result2, result)
		}

		// Verify that the prettified result is successfully parsed into the same string as the original query
		e, err := Parse(s)
		if err != nil {
			t.Fatalf("cannot parse original string: %s", err)
		}
		eResult, err := Parse(result)
		if err != nil {
			t.Fatalf("cannot parse prettified result: %s", err)
		}
		sExpected := e.AppendString(nil)
		sGot := eResult.AppendString(nil)
		if string(sExpected) != string(sGot) {
			t.Fatalf("unexpected prettified string after parsing;\ngot\n%s\nwant\n%s", sGot, sExpected)
		}
	}
	same := func(s string) {
		t.Helper()
		another(s, s)
	}

	// Verify that short queries remain single-line
	same(`foo`)
	same(`foo{bar="baz"}`)
	same(`foo{bar="baz",x="y" or q="w",r="t"}`)
	same(`foo{bar="baz"} + rate(x{y="x"}[5m] offset 1h)`)

	// Verify that empty label filters aren't added to long metric names
	same(`foobar_baz:namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate`)
	another(`foobar_baz:namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate{}`,
		`foobar_baz:namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate`)

	// Verify that long label filters are split into multiple lines
	another(`process_cpu_seconds_total{foo="bar",xjljljlkjopiwererrewre="asdfdsfdsfsdfdsfjkljlk"}`,
		`process_cpu_seconds_total{
  foo="bar",xjljljlkjopiwererrewre="asdfdsfdsfsdfdsfjkljlk"
}`)
	another(`process_cpu_seconds_total{foo="bar",xjljljlkjopiwererrewre="asdfdsfdsfsdfdsfjkljlk",very_long_label_aaaaaaaaaaaaaaa="fdsfdsffdsfs"}`,
		`process_cpu_seconds_total{
  foo="bar",
  xjljljlkjopiwererrewre="asdfdsfdsfsdfdsfjkljlk",
  very_long_label_aaaaaaaaaaaaaaa="fdsfdsffdsfs"
}`)
	another(`{foo="bar",xjljljlkjopiwererrewre="asdfdsfdsfsdfdsfjkljlk",very_long_label_aaaaaaaaaaaaaaa="fdsfdsffdsfs"}`,
		`{
  foo="bar",
  xjljljlkjopiwererrewre="asdfdsfdsfsdfdsfjkljlk",
  very_long_label_aaaaaaaaaaaaaaa="fdsfdsffdsfs"
}`)
	another(`process_cpu_seconds_total{instance="foobar-baz",job="job1234567" or instance="lkjlkjlkjlkjlkjlkjlkjlkjlkjlk",job="lkjljlkjalkadsfdsffdsfdsfd",
		some_very_long_label="very_very_very_long_value_12397787_dfdfdfsds_dsffdfsf"}`,
		`process_cpu_seconds_total{
  instance="foobar-baz",job="job1234567"
    or
  instance="lkjlkjlkjlkjlkjlkjlkjlkjlkjlk",
  job="lkjljlkjalkadsfdsffdsfdsfd",
  some_very_long_label="very_very_very_long_value_12397787_dfdfdfsds_dsffdfsf"
}`)

	// Verify that long binary operations are split into multiple lines
	another(`(sum(rate(process_cpu_seconds_total{instance="foo",job="bar"}[5m] offset 1h @ start())) by (x) / on(x) group_right(y) prefix "x" sum(rate(node_cpu_seconds_total{mode!="idle"}[5m]) keep_metric_names)) keep_metric_names`,
		`(
  sum(
    rate(
      process_cpu_seconds_total{instance="foo",job="bar"}[5m] offset 1h @ start()
    )
  ) by(x)
    / on(x) group_right(y) prefix "x"
  sum(rate(node_cpu_seconds_total{mode!="idle"}[5m]) keep_metric_names)
) keep_metric_names`)

	another(`process_cpu_seconds_total{aaaaaaaaaaaaaaaaaa="bbbbbb"} offset 5m + (rate(xxxxxxxxxxxxxxxx{yyyyyyyy="aaaaaaa"}) keep_metric_names)`,
		`(process_cpu_seconds_total{aaaaaaaaaaaaaaaaaa="bbbbbb"} offset 5m)
  +
(rate(xxxxxxxxxxxxxxxx{yyyyyyyy="aaaaaaa"}) keep_metric_names)`)
	another(`process_cpu_seconds_total{aaaaaaaaaaaaaaaaaa="bbbbbb",cccccccccccccccccccccc!~"ddddddddddddddddddddddd"} offset 5m + (rate(xxxxxxxxxxxxxxxx{yyyyyyyy="aaaaaaa"}) keep_metric_names)`,
		`(
  process_cpu_seconds_total{
    aaaaaaaaaaaaaaaaaa="bbbbbb",
    cccccccccccccccccccccc!~"ddddddddddddddddddddddd"
  } offset 5m
)
  +
(rate(xxxxxxxxxxxxxxxx{yyyyyyyy="aaaaaaa"}) keep_metric_names)`)

	// Verify that long rollup expression is properly split into multiple lines
	another(`process_cpu_seconds_total{foo="bar",aaaaaaaaaaaaaaaaaaaaaaaa="bbbbbbbbbbbbbbbbbbbb",c="dddddddddddd"}[5m:3s] offset 5h3m @ 12345`,
		`process_cpu_seconds_total{
  foo="bar",aaaaaaaaaaaaaaaaaaaaaaaa="bbbbbbbbbbbbbbbbbbbb",c="dddddddddddd"
}[5m:3s] offset 5h3m @ 12345`)
	another(`process_cpu_seconds_total{foo="bar",aaaaaaaaaaaaaaaaaaaaaaaa="bbbbbbbbbbbbbbbbbbbb",ccccccccccccccc="dddddddddddd"}[5m:3s] offset 5h3m @ 12345`,
		`process_cpu_seconds_total{
  foo="bar",
  aaaaaaaaaaaaaaaaaaaaaaaa="bbbbbbbbbbbbbbbbbbbb",
  ccccccccccccccc="dddddddddddd"
}[5m:3s] offset 5h3m @ 12345`)

	// Verify that aggregate expression is properly split into multiple lines
	another(`sum without(x,y) (process_cpu_seconds_total{foo="bar",aaaaaaaaaaaaaaaaaaaaaaaa="bbbbbbbbbbbbbbbbbbbb",c="dddddddddddd"}[5m:3s] offset 5h3m @ 12345)`,
		`sum(
  process_cpu_seconds_total{
    foo="bar",aaaaaaaaaaaaaaaaaaaaaaaa="bbbbbbbbbbbbbbbbbbbb",c="dddddddddddd"
  }[5m:3s] offset 5h3m @ 12345
) without(x,y)`)

	// Verify that an ordinary function args are split into multiple lines
	another(`clamp_min(process_cpu_seconds_total{aaaaaaaaaaaaaaaaaaaaaaaaa="bbbb",cccccc="dddd",ppppppppppppppppppppppppp=~"xxxxxxx"}, 123, "456")`,
		`clamp_min(
  process_cpu_seconds_total{
    aaaaaaaaaaaaaaaaaaaaaaaaa="bbbb",
    cccccc="dddd",
    ppppppppppppppppppppppppp=~"xxxxxxx"
  },
  123,
  "456"
)`)

	// Verify how prettifier works with very long string
	same(`"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`)

	// Verify short WITH expressions
	same(`WITH (x = {a="b"}) x`)
	same(`WITH (x = {a="b"}) {x,y="z"}`)
	same(`WITH (x = {a="b"}) {y="z",x}`)
	same(`WITH (x = {a="b"}) x{y="z"}`)
	same(`WITH (f(s) = {x="foo"+s+"bar"}) f("x")`)

	// Verify long WITH expressions
	another(`with (f(a)=foo{a,qw="ert"}, x = process_cpu_seconds_total{aaaaaaaaaaaaaaaaaaaaaaaaa="bbbb",cccccc="dddd",ppppppppppppppppppppppppp=~"xxxxxxx"}, y=f({ab="cde"}) + efg{h="j"}) x + sum(y)`,
	`WITH (
  f(a) = foo{a,qw="ert"},
  x = (
    process_cpu_seconds_total{
      aaaaaaaaaaaaaaaaaaaaaaaaa="bbbb",
      cccccc="dddd",
      ppppppppppppppppppppppppp=~"xxxxxxx"
    }
  ),
  y = f({ab="cde"}) + efg{h="j"},
)
x + sum(y)`)

	// Embedded WITH expressions
	another(`with ( x = (a{b="c"})  + (with (q=we{rt="z"}) q)) ((abc) / x) + (with(rt=234+234) 2*rt+ poasdfklkjlkjfdsfjklfdfdsfdsfddfsfd)`,
	`WITH (
  x = a{b="c"} + WITH (q = we{rt="z"}) q,
)
(abc / x) + WITH (rt = 234 + 234) (2 * rt) + poasdfklkjlkjfdsfjklfdfdsfdsfddfsfd`)
}
