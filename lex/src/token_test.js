const {block, inline} = require('./rules.js');
const Lexer = require('./Lexer.js');
const {defaults} = require('./defaults.js');
const {Logger} = require(`./log.js`);

const log = new Logger({shortTime: true});

const text = `| AA | BB | CC |
| -- | -- | -- |
| 1 | 2 | 3, **oo** | 
| - 4, x | 5 | 6 *ss*  |
`;
const lex = new Lexer(defaults);
const result = lex.lex(text);

const src = '```'+`

`+'```';