// Generated automatically by nearley, version 2.20.1
// http://github.com/Hardmath123/nearley
function id(x) { return x[0]; }

  import {
    RillTimeAnchor,
    RillTime,
    newSignedRillTimeGrain,
  } from "./RillTime.ts"
let Lexer = undefined;
let ParserRules = [
    {"name": "_$ebnf$1", "symbols": []},
    {"name": "_$ebnf$1", "symbols": ["_$ebnf$1", "wschar"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "_", "symbols": ["_$ebnf$1"], "postprocess": function(d) {return null;}},
    {"name": "__$ebnf$1", "symbols": ["wschar"]},
    {"name": "__$ebnf$1", "symbols": ["__$ebnf$1", "wschar"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "__", "symbols": ["__$ebnf$1"], "postprocess": function(d) {return null;}},
    {"name": "wschar", "symbols": [/[ \t\n\v\f]/], "postprocess": id},
    {"name": "dqstring$ebnf$1", "symbols": []},
    {"name": "dqstring$ebnf$1", "symbols": ["dqstring$ebnf$1", "dstrchar"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "dqstring", "symbols": [{"literal":"\""}, "dqstring$ebnf$1", {"literal":"\""}], "postprocess": function(d) {return d[1].join(""); }},
    {"name": "sqstring$ebnf$1", "symbols": []},
    {"name": "sqstring$ebnf$1", "symbols": ["sqstring$ebnf$1", "sstrchar"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "sqstring", "symbols": [{"literal":"'"}, "sqstring$ebnf$1", {"literal":"'"}], "postprocess": function(d) {return d[1].join(""); }},
    {"name": "btstring$ebnf$1", "symbols": []},
    {"name": "btstring$ebnf$1", "symbols": ["btstring$ebnf$1", /[^`]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "btstring", "symbols": [{"literal":"`"}, "btstring$ebnf$1", {"literal":"`"}], "postprocess": function(d) {return d[1].join(""); }},
    {"name": "dstrchar", "symbols": [/[^\\"\n]/], "postprocess": id},
    {"name": "dstrchar", "symbols": [{"literal":"\\"}, "strescape"], "postprocess": 
        function(d) {
            return JSON.parse("\""+d.join("")+"\"");
        }
        },
    {"name": "sstrchar", "symbols": [/[^\\'\n]/], "postprocess": id},
    {"name": "sstrchar", "symbols": [{"literal":"\\"}, "strescape"], "postprocess": function(d) { return JSON.parse("\""+d.join("")+"\""); }},
    {"name": "sstrchar$string$1", "symbols": [{"literal":"\\"}, {"literal":"'"}], "postprocess": function joiner(d) {return d.join('');}},
    {"name": "sstrchar", "symbols": ["sstrchar$string$1"], "postprocess": function(d) {return "'"; }},
    {"name": "strescape", "symbols": [/["\\/bfnrt]/], "postprocess": id},
    {"name": "strescape", "symbols": [{"literal":"u"}, /[a-fA-F0-9]/, /[a-fA-F0-9]/, /[a-fA-F0-9]/, /[a-fA-F0-9]/], "postprocess": 
        function(d) {
            return d.join("");
        }
        },
    {"name": "rill_time", "symbols": ["time_anchor_part"], "postprocess": ([{ start, end }]) => new RillTime(start, end)},
    {"name": "rill_time", "symbols": ["time_anchor_part", "_", "grain_and_at_part"], "postprocess": ([{ start, end }, , { grain, modifier }]) => new RillTime(start, end, grain, modifier)},
    {"name": "time_anchor_part", "symbols": ["time_anchor", "_", {"literal":","}, "_", "time_anchor"], "postprocess": ([start, , , , end]) => ({ start, end })},
    {"name": "time_anchor_part", "symbols": ["time_anchor"], "postprocess": ([start]) => ({ start })},
    {"name": "time_anchor", "symbols": ["time_anchor_offset", "_", {"literal":"/"}, "_", "grain_modifier", "_", "time_anchor_offset"], "postprocess": ([mod, , , , truncate, , offset]) => mod.withOffset(offset).withTruncate(truncate)},
    {"name": "time_anchor", "symbols": ["time_anchor_offset", "_", "time_anchor_offset"], "postprocess": ([rillTime, , offset]) => rillTime.withOffset(offset)},
    {"name": "time_anchor", "symbols": ["time_anchor_offset", "_", {"literal":"/"}, "_", "grain_modifier"], "postprocess": ([mod, , , , truncate]) => mod.withTruncate(truncate)},
    {"name": "time_anchor", "symbols": ["time_anchor_offset"], "postprocess": id},
    {"name": "time_anchor_offset$string$1", "symbols": [{"literal":"n"}, {"literal":"o"}, {"literal":"w"}], "postprocess": function joiner(d) {return d.join('');}},
    {"name": "time_anchor_offset", "symbols": ["time_anchor_offset$string$1"], "postprocess": () => RillTimeAnchor.now()},
    {"name": "time_anchor_offset$string$2", "symbols": [{"literal":"e"}, {"literal":"a"}, {"literal":"r"}, {"literal":"l"}, {"literal":"i"}, {"literal":"e"}, {"literal":"s"}, {"literal":"t"}], "postprocess": function joiner(d) {return d.join('');}},
    {"name": "time_anchor_offset", "symbols": ["time_anchor_offset$string$2"], "postprocess": () => RillTimeAnchor.earliest()},
    {"name": "time_anchor_offset$string$3", "symbols": [{"literal":"l"}, {"literal":"a"}, {"literal":"t"}, {"literal":"e"}, {"literal":"s"}, {"literal":"t"}], "postprocess": function joiner(d) {return d.join('');}},
    {"name": "time_anchor_offset", "symbols": ["time_anchor_offset$string$3"], "postprocess": () => RillTimeAnchor.latest()},
    {"name": "time_anchor_offset$string$4", "symbols": [{"literal":"w"}, {"literal":"a"}, {"literal":"t"}, {"literal":"e"}, {"literal":"r"}, {"literal":"m"}, {"literal":"a"}, {"literal":"r"}, {"literal":"k"}], "postprocess": function joiner(d) {return d.join('');}},
    {"name": "time_anchor_offset", "symbols": ["time_anchor_offset$string$4"], "postprocess": () => RillTimeAnchor.latest()},
    {"name": "time_anchor_offset", "symbols": ["abs_time"], "postprocess": ([absTime]) => RillTimeAnchor.absolute(absTime)},
    {"name": "time_anchor_offset", "symbols": ["grain_modifier"], "postprocess": ([grain]) => RillTimeAnchor.relative(grain)},
    {"name": "grain_and_at_part", "symbols": [{"literal":":"}, "_", "range_grain_modifier", "_", {"literal":"@"}, "_", "at_modifiers"], "postprocess": ([, , grain, , , , modifier]) => ({ grain, modifier })},
    {"name": "grain_and_at_part", "symbols": [{"literal":":"}, "_", "range_grain_modifier"], "postprocess": ([, , grain]) => ({ grain })},
    {"name": "grain_and_at_part", "symbols": [{"literal":"@"}, "_", "at_modifiers"], "postprocess": ([, , modifier]) => ({ modifier })},
    {"name": "range_grain_modifier", "symbols": ["grain"], "postprocess": ([grain]) => ({ grain, isComplete: false })},
    {"name": "range_grain_modifier", "symbols": [{"literal":"|"}, "_", "grain", "_", {"literal":"|"}], "postprocess": ([, ,grain]) => ({ grain, isComplete: true })},
    {"name": "at_modifiers", "symbols": ["time_anchor_offset"], "postprocess": ([grain]) => ({ at: grain })},
    {"name": "at_modifiers", "symbols": ["timezone_modifier"], "postprocess": ([timeZone]) => ({ timeZone })},
    {"name": "at_modifiers", "symbols": ["time_anchor_offset", "_", "timezone_modifier"], "postprocess": ([grain, , timeZone]) => ({ at: grain, timeZone })},
    {"name": "grain_modifier", "symbols": ["grain"], "postprocess": ([grain]) => ({ count: 0, grain })},
    {"name": "grain_modifier", "symbols": ["sign", "_", "grain"], "postprocess": ([sign, , grain]) => newSignedRillTimeGrain(grain, sign, 0)},
    {"name": "grain_modifier", "symbols": ["num", "grain"], "postprocess": ([count, grain]) => ({ count, grain })},
    {"name": "grain_modifier", "symbols": ["sign", "_", "num", "grain"], "postprocess": ([sign, , count, grain]) => newSignedRillTimeGrain(grain, sign, count)},
    {"name": "abs_time", "symbols": [/[\d]/, /[\d]/, /[\d]/, /[\d]/, /[\-]/, /[\d]/, /[\d]/, /[\-]/, /[\d]/, /[\d]/, "_", /[\d]/, /[\d]/, /[:]/, /[\d]/, /[\d]/], "postprocess": (args) => args.join("")},
    {"name": "abs_time", "symbols": [/[\d]/, /[\d]/, /[\d]/, /[\d]/, /[\-]/, /[\d]/, /[\d]/, /[\-]/, /[\d]/, /[\d]/], "postprocess": (args) => args.join("")},
    {"name": "timezone_modifier$ebnf$1", "symbols": [/[^}]/]},
    {"name": "timezone_modifier$ebnf$1", "symbols": ["timezone_modifier$ebnf$1", /[^}]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "timezone_modifier", "symbols": [{"literal":"{"}, "_", "timezone_modifier$ebnf$1", "_", {"literal":"}"}], "postprocess": ([, , tz]) => tz.join("")},
    {"name": "sign", "symbols": [/[+-]/], "postprocess": id},
    {"name": "num$ebnf$1", "symbols": [/[0-9]/]},
    {"name": "num$ebnf$1", "symbols": ["num$ebnf$1", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "num", "symbols": ["num$ebnf$1"], "postprocess": (args) => Number(args.join(""))},
    {"name": "grain", "symbols": [/[sSmhHdDwWqQMyY]/], "postprocess": id}
];
let ParserStart = "rill_time";
export default { Lexer, ParserRules, ParserStart };
