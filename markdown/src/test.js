function merge(obj) {
    let i = 1,
        target,
        key;

    for (; i < arguments.length; i++) {
        console.log(arguments[0], arguments[1]);
        target = arguments[i];
        for (key in target) {
            console.log();
            if (Object.prototype.hasOwnProperty.call(target, key)) {
                obj[key] = target[key];
            }
        }
    }

    return obj;
}

console.log(merge({"aaa": "vv"}, {
    "aaa": "vv1",
    "www": function () {
        console.log("wwww")
    }
}));