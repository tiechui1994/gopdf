/**
 * Helpers
 */

const caret = /(^|[^\[])\^/g;

function edit(regex, opt) {
    regex = regex.source || regex;
    opt = opt || '';
    const obj = {
        replace: (name, val) => {
            val = val.source || val;
            val = val.replace(caret, '$1');
            regex = regex.replace(name, val);
            return obj;
        },
        getRegex: () => {
            return new RegExp(regex, opt);
        }
    };
    return obj;
}

const noopTest = {
    exec: function noopTest() {
    }
};

function merge(obj) {
    let i = 1,
        target,
        key;

    for (; i < arguments.length; i++) {
        target = arguments[i];
        for (key in target) {
            if (Object.prototype.hasOwnProperty.call(target, key)) {
                obj[key] = target[key];
            }
        }
    }

    return obj;
}

function splitCells(tableRow, count) {
    // ensure that every cell-delimiting pipe has a space
    // before it to distinguish it from an escaped pipe
    const row = tableRow.replace(/\|/g, (match, offset, str) => {
            let escaped = false,
                curr = offset;
            while (--curr >= 0 && str[curr] === '\\') escaped = !escaped;
            if (escaped) {
                // odd number of slashes means | is escaped
                // so we leave it alone
                return '|';
            } else {
                // add space before unescaped |
                return ' |';
            }
        }),
        cells = row.split(/ \|/);
    let i = 0;

    if (cells.length > count) {
        cells.splice(count);
    } else {
        while (cells.length < count) cells.push('');
    }

    for (; i < cells.length; i++) {
        // leading or trailing whitespace is ignored per the gfm spec
        cells[i] = cells[i].trim().replace(/\\\|/g, '|');
    }
    return cells;
}

// Remove trailing 'c's. Equivalent to str.replace(/c*$/, '').
// /c*$/ is vulnerable to REDOS.
// invert: Remove suffix of non-c chars instead. Default falsey.
function rtrim(str, c, invert) {
    const l = str.length;
    if (l === 0) {
        return '';
    }

    // Length of suffix matching the invert condition.
    let suffLen = 0;

    // Step left until we fail to match the invert condition.
    while (suffLen < l) {
        const currChar = str.charAt(l - suffLen - 1);
        if (currChar === c && !invert) {
            suffLen++;
        } else if (currChar !== c && invert) {
            suffLen++;
        } else {
            break;
        }
    }

    return str.substr(0, l - suffLen);
}

function findClosingBracket(str, b) {
    if (str.indexOf(b[1]) === -1) {
        return -1;
    }
    const l = str.length;
    let level = 0,
        i = 0;
    for (; i < l; i++) {
        if (str[i] === '\\') {
            i++;
        } else if (str[i] === b[0]) {
            level++;
        } else if (str[i] === b[1]) {
            level--;
            if (level < 0) {
                return i;
            }
        }
    }
    return -1;
}


module.exports = {
    edit,
    noopTest,
    merge,
    splitCells,
    rtrim,
    findClosingBracket,
};
