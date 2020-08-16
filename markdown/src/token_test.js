const {block, inline} = require('./rules.js');
const Lexer = require('./Lexer.js');
const {defaults} = require('./defaults.js');
const {Logger} = require(`./log.js`);

const log = new Logger({shortTime: true});

const text = `
Markdown Quick Reference
========================

This guide is a very brief overview, with examples, of the syntax that [Markdown] supports. It is itself written in Markdown and you can copy the samples over to the left-hand pane for experimentation. It is shown as *text* and not *rendered HTML*.

[Markdown]: http://daringfireball.net/projects/markdown/


Simple Text Formatting
======================

First thing is first. You can use *stars* or _underscores_ for italics. **Double stars** and __double underscores__ for bold. ***Three together*** for ___both___.

Paragraphs are pretty easy too. Just have a blank line between chunks of text.
`;
const lex = new Lexer(defaults);
const result = lex.lex(text);
log.warn(result);