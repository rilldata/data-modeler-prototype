@preprocessor esmodule
@builtin "whitespace.ne"
@builtin "string.ne"

@{%
  import {
    RillTimeAnchor,
    RillTime,
    newSignedRillTimeGrain,
  } from "./RillTime.ts"
%}

rill_time => time_anchor_part                     {% ([{ start, end }]) => new RillTime(start, end) %}
           | time_anchor_part _ grain_and_at_part {% ([{ start, end }, , { grain, modifier }]) => new RillTime(start, end, grain, modifier) %}

time_anchor_part => time_anchor _ "," _ time_anchor {% ([start, , , , end]) => ({ start, end }) %}
                  | time_anchor                     {% ([start]) => ({ start }) %}

time_anchor => time_anchor_offset _ "/" _ grain_modifier _ time_anchor_offset {% ([mod, , , , truncate, , offset]) => mod.withOffset(offset).withTruncate(truncate) %}
             | time_anchor_offset _ time_anchor_offset                        {% ([rillTime, , offset]) => rillTime.withOffset(offset) %}
             | time_anchor_offset _ "/" _ grain_modifier                      {% ([mod, , , , truncate]) => mod.withTruncate(truncate) %}
             | time_anchor_offset                                             {% id %}

time_anchor_offset => "now"          {% () => RillTimeAnchor.now() %}
                    | "earliest"     {% () => RillTimeAnchor.earliest() %}
                    | "latest"       {% () => RillTimeAnchor.latest() %}
                    | "watermark"    {% () => RillTimeAnchor.latest() %}
                    | abs_time       {% ([absTime]) => RillTimeAnchor.absolute(absTime) %}
                    | grain_modifier {% ([grain]) => RillTimeAnchor.relative(grain) %}

grain_and_at_part => ":" _ range_grain_modifier _ "@" _ at_modifiers {% ([, , grain, , , , modifier]) => ({ grain, modifier }) %}
                   | ":" _ range_grain_modifier                      {% ([, , grain]) => ({ grain }) %}
                   | "@" _ at_modifiers                              {% ([, , modifier]) => ({ modifier }) %}

range_grain_modifier => grain             {% ([grain]) => ({ grain, isComplete: false }) %}
                      | "|" _ grain _ "|" {% ([, ,grain]) => ({ grain, isComplete: true }) %}

at_modifiers => time_anchor_offset                     {% ([grain]) => ({ at: grain }) %}
              | timezone_modifier                      {% ([timeZone]) => ({ timeZone }) %}
              | time_anchor_offset _ timezone_modifier {% ([grain, , timeZone]) => ({ at: grain, timeZone }) %}

grain_modifier => grain            {% ([grain]) => ({ count: 0, grain }) %}
                | sign _ grain     {% ([sign, , grain]) => newSignedRillTimeGrain(grain, sign, 0) %}
                | num grain        {% ([count, grain]) => ({ count, grain }) %}
                | sign _ num grain {% ([sign, , count, grain]) => newSignedRillTimeGrain(grain, sign, count) %}

abs_time => [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] _ [\d] [\d] [:] [\d] [\d] {% (args) => args.join("") %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] {% (args) => args.join("") %}

timezone_modifier => "{" _ [^}]:+ _ "}" {% ([, , tz]) => tz.join("") %}

sign => [+-] {% id %}

num => [0-9]:+ {% (args) => Number(args.join("")) %}

grain => [sSmhHdDwWqQMyY] {% id %}
