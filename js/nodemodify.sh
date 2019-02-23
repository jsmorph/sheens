#!/bin/bash

# This goofy script attempts to generate a Node module based on the
# Javascript sources in 'js/'.
#
# Uses https://www.npmjs.com/package/safe-eval.
#
# After running this script, try
#
#   sudo npm link node-sheens
#
# Then
#
#   npm install safe-eval
#
# BUT 'safe-eval' is of course not that safe.  See
# https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-16088 and
# probably elsewhere.
#
#   const SHEENS = require('sheens');
#

set -e

TARGET=node-sheens
mkdir -p $TARGET

cat<<EOF > $TARGET/index.js
function print(x) {
    console.log(x);
}

var safeEval = require('safe-eval');

var sandbox = function(code) {
    return safeEval(code);
}
EOF

for F in prof match sandbox step; do 
    cat $F.js >> $TARGET/index.js
done

cat<<EOF >> $TARGET/index.js
exports.step = step;
exports.walk = walk;
exports.match = match;
exports.action = sandboxedAction;
exports.times = Times;
EOF

cat<<EOF > $TARGET/package.json
{
  "name": "sheens",
  "version": "1.0.0",
  "description": "Sheens https://github.com/jsmorph/sheens",
  "main": "index.js",
  "scripts": {
    "test": "echo \"Everything is perfect\""
  },
  "author": "",
  "license": "ISC"
}
EOF

echo "Wrote $TARGET"

