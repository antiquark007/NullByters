class SuccessManager {
    constructor() {
        this.wipeResult = null;
        this.certificate = null;
        this.lastCertPath = null;
        this.init();
    }

    init() {
        // Get wipe result from sessionStorage and localStorage
        const logPath = sessionStorage.getItem('lastWipeLog');
        const devicePath = sessionStorage.getItem('lastDevicePath');
        const method = sessionStorage.getItem('lastWipeMethod');
        
        // Try to get enhanced wipe result from localStorage (from progress.js)
        this.wipeResult = JSON.parse(localStorage.getItem('wipeResult') || '{}');
        
        // Fallback to session data if localStorage is empty
        if (!this.wipeResult.success && logPath) {
            this.wipeResult = {
                success: true,
                logPath: logPath,
                devicePath: devicePath,
                method: method,
                timestamp: new Date().toISOString()
            };
        }

        this.setupElements();
        this.displayWipeResults();
        this.setupEventHandlers();
        
        // Auto-generate certificate if wipe was successful
        if (this.wipeResult.success && this.wipeResult.logPath) {
            setTimeout(() => this.generateCertificate(), 1000);
        }
    }

    setupElements() {
        // Get existing elements
        this.summaryArea = document.getElementById('summaryArea');
        this.genBtn = document.getElementById('genCert');
        this.viewJsonBtn = document.getElementById('viewJson');
        this.exportBtn = document.getElementById('exportBtn');
        this.homeBtn = document.getElementById('home');
        this.certStatus = document.getElementById('certStatus');

        // Create additional elements for enhanced features
        if (!document.getElementById('deviceInfo')) {
            const deviceInfoDiv = document.createElement('div');
            deviceInfoDiv.id = 'deviceInfo';
            deviceInfoDiv.className = 'device-info-section';
            this.summaryArea.parentNode.insertBefore(deviceInfoDiv, this.summaryArea);
        }

        if (!document.getElementById('certificateDetails')) {
            const certDetailsDiv = document.createElement('div');
            certDetailsDiv.id = 'certificateDetails';
            certDetailsDiv.className = 'certificate-details';
            this.certStatus.parentNode.appendChild(certDetailsDiv);
        }

        if (!document.getElementById('verificationSection')) {
            const verifyDiv = document.createElement('div');
            verifyDiv.id = 'verificationSection';
            verifyDiv.className = 'verification-section';
            this.certStatus.parentNode.appendChild(verifyDiv);
        }
    }

    displayWipeResults() {
        const deviceInfo = this.wipeResult.deviceInfo || {};
        
        // Update device information
        document.getElementById('deviceInfo').innerHTML = `
            <h3>📋 Wipe Operation Summary</h3>
            <div class="info-grid">
                <div class="info-item">
                    <strong>Device:</strong> ${deviceInfo.name || 'Unknown Device'}
                </div>
                <div class="info-item">
                    <strong>Path:</strong> ${this.wipeResult.devicePath || 'N/A'}
                </div>
                <div class="info-item">
                    <strong>Method:</strong> ${(this.wipeResult.method || 'unknown').toUpperCase()}
                </div>
                <div class="info-item">
                    <strong>Size:</strong> ${deviceInfo.size_gb ? deviceInfo.size_gb.toFixed(1) + ' GB' : 'Unknown'}
                </div>
                <div class="info-item">
                    <strong>Serial:</strong> ${deviceInfo.serial || 'N/A'}
                </div>
                <div class="info-item">
                    <strong>Completed:</strong> ${new Date(this.wipeResult.timestamp).toLocaleString()}
                </div>
            </div>
        `;

        // Update summary area
        this.summaryArea.innerHTML = `
            <div class="wipe-status ${this.wipeResult.success ? 'success' : 'error'}">
                <div class="status-icon">${this.wipeResult.success ? '✅' : '❌'}</div>
                <div class="status-text">
                    <h4>${this.wipeResult.success ? 'Secure Wipe Completed Successfully' : 'Wipe Operation Failed'}</h4>
                    <div class="small-muted">Log Path: ${this.wipeResult.logPath || 'N/A'}</div>
                    ${!this.wipeResult.success ? `<div class="error-details">Error: ${this.wipeResult.error || 'Unknown error'}</div>` : ''}
                </div>
            </div>
        `;
    }

    setupEventHandlers() {
        // Generate Certificate Button
        this.genBtn.addEventListener('click', () => this.generateCertificate());

        // View JSON Button - Updated API
        this.viewJsonBtn.addEventListener('click', async () => {
            if (!this.lastCertPath) {
                alert('No certificate found. Please generate a certificate first.');
                return;
            }
            
            try {
                const saveRes = await window.api.showSaveDialog({ 
                    title: 'Save certificate JSON', 
                    defaultPath: `certificate_${Date.now()}.json`,
                    filters: [
                        { name: 'JSON Files', extensions: ['json'] },
                        { name: 'All Files', extensions: ['*'] }
                    ]
                });
                
                if (!saveRes.canceled) {
                    const copyRes = await window.api.copyFile({ 
                        source: this.lastCertPath, 
                        destination: saveRes.filePath 
                    });
                    
                    if (copyRes.success) {
                        this.showMessage('✅ Certificate saved successfully!', 'success');
                    } else {
                        this.showMessage(`❌ Save failed: ${copyRes.error}`, 'error');
                    }
                }
            } catch (error) {
                this.showMessage(`❌ Error: ${error.message}`, 'error');
            }
        });

        // Export Button - Enhanced functionality
        this.exportBtn.addEventListener('click', async () => {
            if (!this.certificate) {
                alert('No certificate to export. Please generate a certificate first.');
                return;
            }

            try {
                const saveRes = await window.api.showSaveDialog({ 
                    title: 'Export certificate package', 
                    defaultPath: `certificate_package_${Date.now()}.json`,
                    filters: [
                        { name: 'JSON Files', extensions: ['json'] },
                        { name: 'PDF Files', extensions: ['pdf'] },
                        { name: 'All Files', extensions: ['*'] }
                    ]
                });
                
                if (!saveRes.canceled) {
                    // Export both JSON and PDF if available
                    const jsonRes = await window.api.copyFile({ 
                        source: this.certificate.jsonPath, 
                        destination: saveRes.filePath 
                    });
                    
                    if (this.certificate.pdfPath) {
                        const pdfPath = saveRes.filePath.replace(/\.json$/, '.pdf');
                        await window.api.copyFile({ 
                            source: this.certificate.pdfPath, 
                            destination: pdfPath 
                        });
                    }
                    
                    if (jsonRes.success) {
                        this.showMessage('✅ Certificate package exported successfully!', 'success');
                    } else {
                        this.showMessage(`❌ Export failed: ${jsonRes.error}`, 'error');
                    }
                }
            } catch (error) {
                this.showMessage(`❌ Export error: ${error.message}`, 'error');
            }
        });

        // Home Button
        this.homeBtn.addEventListener('click', () => {
            // Clean up session data
            sessionStorage.removeItem('lastWipeLog');
            sessionStorage.removeItem('lastDevicePath');
            sessionStorage.removeItem('lastWipeMethod');
            localStorage.removeItem('wipeResult');
            
            window.api.loadPage('landing.html');
        });

        // Add new buttons for enhanced features
        this.addEnhancedButtons();
    }

    addEnhancedButtons() {
        const buttonContainer = this.exportBtn.parentNode;
        
        // Add Verify Certificate button
        if (!document.getElementById('verifyBtn')) {
            const verifyBtn = document.createElement('button');
            verifyBtn.id = 'verifyBtn';
            verifyBtn.className = 'btn btn-info';
            verifyBtn.textContent = '🔍 Verify Certificate';
            verifyBtn.disabled = true;
            verifyBtn.addEventListener('click', () => this.verifyCertificate());
            buttonContainer.appendChild(verifyBtn);
        }

        // Add Download PDF button
        if (!document.getElementById('downloadPdfBtn')) {
            const downloadPdfBtn = document.createElement('button');
            downloadPdfBtn.id = 'downloadPdfBtn';
            downloadPdfBtn.className = 'btn btn-secondary';
            downloadPdfBtn.textContent = '📄 Download PDF';
            downloadPdfBtn.disabled = true;
            downloadPdfBtn.addEventListener('click', () => this.downloadPdf());
            buttonContainer.appendChild(downloadPdfBtn);
        }

        // Add Show QR Code button
        if (!document.getElementById('showQrBtn')) {
            const showQrBtn = document.createElement('button');
            showQrBtn.id = 'showQrBtn';
            showQrBtn.className = 'btn btn-info';
            showQrBtn.textContent = '📱 Show QR Code';
            showQrBtn.disabled = true;
            showQrBtn.addEventListener('click', () => this.toggleQrCode());
            buttonContainer.appendChild(showQrBtn);
        }
    }

    async generateCertificate() {
        if (!this.wipeResult.logPath) {
            this.showMessage('❌ No wipe log found. Cannot generate certificate.', 'error');
            return;
        }

        this.genBtn.disabled = true;
        this.certStatus.innerText = '🔄 Generating certificate...';
        
        try {
            const certArgs = {
                logPath: this.wipeResult.logPath,
                outJson: `cert_${Date.now()}.json`,
                outPdf: `cert_${Date.now()}.pdf`,
                deviceInfo: {
                    path: this.wipeResult.devicePath,
                    name: this.wipeResult.deviceInfo?.name || 'Unknown Device',
                    method: this.wipeResult.method,
                    size_gb: this.wipeResult.deviceInfo?.size_gb,
                    serial: this.wipeResult.deviceInfo?.serial
                }
            };
            
            console.log('[DEBUG] Generating certificate with args:', certArgs);
            
            const result = await window.api.generateCert(certArgs);
            console.log('[DEBUG] Certificate result:', result);
            
            if (result.success) {
                this.certificate = result;
                this.lastCertPath = result.jsonPath;
                
                const mockBadge = result.mock ? ' (Demo Mode)' : '';
                this.certStatus.innerText = `✅ Certificate generated successfully${mockBadge}`;
                
                this.displayCertificateDetails();
                this.enableCertificateButtons();
                
                this.showMessage('🎉 Compliance certificate generated!', 'success');
            } else {
                this.certStatus.innerText = `❌ Generation failed: ${result.error}`;
                this.showMessage(`Certificate generation failed: ${result.error}`, 'error');
            }
            
        } catch (error) {
            console.error('Certificate generation error:', error);
            this.certStatus.innerText = `❌ Error: ${error.message}`;
            this.showMessage(`Error generating certificate: ${error.message}`, 'error');
        } finally {
            this.genBtn.disabled = false;
        }
    }

    displayCertificateDetails() {
        const detailsDiv = document.getElementById('certificateDetails');
        const mockBadge = this.certificate.mock ? '<span class="mock-badge">DEMO MODE</span>' : '';
        
        detailsDiv.innerHTML = `
            <div class="certificate-info">
                <h3>📜 Certificate Details ${mockBadge}</h3>
                <div class="cert-grid">
                    <div class="cert-item">
                        <strong>Certificate ID:</strong>
                        <code>${this.certificate.certificate_id}</code>
                    </div>
                    <div class="cert-item">
                        <strong>JSON Path:</strong>
                        <code>${this.certificate.jsonPath}</code>
                    </div>
                    <div class="cert-item">
                        <strong>PDF Path:</strong>
                        <code>${this.certificate.pdfPath || 'Not generated'}</code>
                    </div>
                    <div class="cert-item">
                        <strong>QR Code:</strong>
                        <span>${this.certificate.qr_data ? 'Available' : 'Not generated'}</span>
                    </div>
                </div>
                
                <div id="qrCodeSection" class="qr-section" style="display: none;">
                    <h4>📱 QR Code for Mobile Verification</h4>
                    <div class="qr-data">
                        <p>Scan with any QR code reader to verify certificate:</p>
                        <code>${this.certificate.qr_data || 'QR data not available'}</code>
                    </div>
                </div>
            </div>
        `;
    }

    enableCertificateButtons() {
        this.viewJsonBtn.disabled = false;
        this.exportBtn.disabled = false;
        document.getElementById('verifyBtn').disabled = false;
        
        if (this.certificate.pdfPath) {
            document.getElementById('downloadPdfBtn').disabled = false;
        }
        
        if (this.certificate.qr_data) {
            document.getElementById('showQrBtn').disabled = false;
        }
    }

    async verifyCertificate() {
        if (!this.certificate) return;
        
        const verifySection = document.getElementById('verificationSection');
        verifySection.innerHTML = '<div class="loading">🔄 Verifying certificate...</div>';
        verifySection.style.display = 'block';
        
        try {
            const result = await window.api.verifyCert({
                certPath: this.certificate.jsonPath,
                qrData: this.certificate.qr_data
            });
            
            if (result.valid || result.verified) {
                verifySection.innerHTML = `
                    <div class="verification-success">
                        ✅ Certificate verification successful!
                        ${result.mock ? '<br><em>(Demo mode verification)</em>' : ''}
                        <br><small>${result.message || 'Certificate is authentic and has not been tampered with.'}</small>
                    </div>
                `;
            } else {
                verifySection.innerHTML = `
                    <div class="verification-error">
                        ❌ Certificate verification failed!
                        <br><small>${result.error || result.message || 'Certificate may be invalid or corrupted.'}</small>
                    </div>
                `;
            }
        } catch (error) {
            verifySection.innerHTML = `
                <div class="verification-error">
                    ❌ Verification error: ${error.message}
                </div>
            `;
        }
    }

    async downloadPdf() {
        if (!this.certificate.pdfPath) return;
        
        try {
            const saveRes = await window.api.showSaveDialog({ 
                title: 'Save PDF certificate', 
                defaultPath: `certificate_${Date.now()}.pdf`,
                filters: [
                    { name: 'PDF Files', extensions: ['pdf'] },
                    { name: 'All Files', extensions: ['*'] }
                ]
            });
            
            if (!saveRes.canceled) {
                const copyRes = await window.api.copyFile({ 
                    source: this.certificate.pdfPath, 
                    destination: saveRes.filePath 
                });
                
                if (copyRes.success) {
                    this.showMessage('✅ PDF certificate saved!', 'success');
                } else {
                    this.showMessage(`❌ PDF save failed: ${copyRes.error}`, 'error');
                }
            }
        } catch (error) {
            this.showMessage(`❌ PDF download error: ${error.message}`, 'error');
        }
    }

    toggleQrCode() {
        const qrSection = document.getElementById('qrCodeSection');
        qrSection.style.display = qrSection.style.display === 'none' ? 'block' : 'none';
    }

    showMessage(message, type = 'info') {
        // Create or update message display
        let messageDiv = document.getElementById('messageDisplay');
        if (!messageDiv) {
            messageDiv = document.createElement('div');
            messageDiv.id = 'messageDisplay';
            messageDiv.className = 'message-display';
            this.certStatus.parentNode.appendChild(messageDiv);
        }
        
        messageDiv.innerHTML = `<div class="message ${type}">${message}</div>`;
        messageDiv.style.display = 'block';
        
        // Auto-hide after 5 seconds
        setTimeout(() => {
            messageDiv.style.display = 'none';
        }, 5000);
    }
}

// Initialize the enhanced success manager
document.addEventListener('DOMContentLoaded', () => {
    new SuccessManager();
});

// Handle navigation from other pages
window.addEventListener('storage', (e) => {
    if (e.key === 'wipeResult' && e.newValue) {
        location.reload();
    }
});