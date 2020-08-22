const fs = require('fs');
const Lexer = require('./Lexer.js');
const {defaults} = require('./defaults.js');

const src = fs.readFileSync('./mark.md');
const tokens = Lexer.lex(src.toString(), defaults);
fs.writeFile("./mark.json", JSON.stringify(tokens), (err) => {
    if (err) {
        console.log("err", err)
    } else {
        console.log("success")
    }
});