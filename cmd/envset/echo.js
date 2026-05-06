console.log('-------- ECHO ---------');
console.log('cmd: "%s"', process.argv.join(' '));

const _title = function(str) {
    str = str
        .replace(/_/g, ' ')
        .toLowerCase();
    return str.charAt(0).toUpperCase() + str.slice(1);
};

Object.keys(process.env).forEach(key => {
    console.log('%s => %s', key, process.env[key]);
});

console.log('Arg:', process.argv[2]);
console.log('Empty:', process.argv[3]);