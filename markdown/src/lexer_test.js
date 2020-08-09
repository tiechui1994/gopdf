const fs = require('fs');
const Lexer = require('./Lexer.js');
const {defaults} = require('./defaults.js');

const src = fs.readFileSync('./mark.md');

const tokens = Lexer.lex(src.toString(), defaults);

const files = JSON.stringify(tokens);

fs.writeFile("./mark.json", files, (err) => {});