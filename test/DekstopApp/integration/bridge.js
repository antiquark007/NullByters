const { spawn } = require('child_process');
const path = require('path');

const wipeScriptPath = path.join(__dirname, '../wipe-tool/wipe.py');

function runWipeScript(deviceType, deviceInfo) {
    return new Promise((resolve, reject) => {
        const pythonProcess = spawn('python3', [wipeScriptPath, deviceType, deviceInfo]);

        pythonProcess.stdout.on('data', (data) => {
            console.log(`Output: ${data}`);
        });

        pythonProcess.stderr.on('data', (data) => {
            console.error(`Error: ${data}`);
        });

        pythonProcess.on('close', (code) => {
            if (code === 0) {
                resolve('Wipe operation completed successfully.');
            } else {
                reject(`Wipe operation failed with code: ${code}`);
            }
        });
    });
}

module.exports = {
    runWipeScript,
};