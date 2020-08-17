const {block, inline} = require('./rules.js');
const Lexer = require('./Lexer.js');
const {defaults} = require('./defaults.js');
const {Logger} = require(`./log.js`);

const log = new Logger({shortTime: true});

const text = `
Email addresses in plain text are not linked: test@example.com.
Email addresses wrapped in angle brackets are linked: <test@example.com>.
They are also obfuscated so that email harvesting spam robots hopefully won not get them.
`;
const lex = new Lexer(defaults);
const result = lex.lex(text);
log.warn(result);