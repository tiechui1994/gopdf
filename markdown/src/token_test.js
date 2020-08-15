const {block, inline} = require('./rules.js');

const fs = require("fs");
const src = fs.readFileSync('./inline.json');
const obj = JSON.parse(src.toString());

for (let key in obj) {
    console.log(key);
    if (inline[key]) {
        const re = new RegExp(obj[key]);
        if (inline[key].toString() == re.toString()) {
            console.log("ok")
        } else {
            console.log(inline[key].toString())
            console.log(re.toString())
        }

        console.log("\n++++++++++++++++++++++++++++++++++\n")
    }
}