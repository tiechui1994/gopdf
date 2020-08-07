const fs = require('fs');
const Lexer = require('./Lexer.js');
const {defaults} = require('./defaults.js');

function log(...message) {
    console.log(message)
}

const src = fs.readFileSync('./mark.md');

const tokens = Lexer.lex(src.toString(), defaults);

log(JSON.stringify(tokens));

