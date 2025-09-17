const { spawn } = require('child_process');
const path = require('path');

const pythonScriptPath = path.join(__dirname, '../wipe-tool/wipe.py');

function runWipeScript(args) {
    return new Promise((resolve, reject) => {
        const process = spawn('python3', [pythonScriptPath, ...args]);

        let output = '';
        let errorOutput = '';

        process.stdout.on('data', (data) => {
            output += data.toString();
        });

        process.stderr.on('data', (data) => {
            errorOutput += data.toString();
        });

        process.on('close', (code) => {
            if (code !== 0) {
                reject(new Error(`Wipe script exited with code ${code}: ${errorOutput}`));
            } else {
                resolve(output);
            }
        });
    });
}

module.exports = {
    runWipeScript
};