// Interactive CRA Risk Assessment Wizard
let currentQuestion = 1;
let totalQuestions = 21;
let riskScore = 0;
let answers = {};

const questions = [
    {
        id: 1,
        icon: '📦',
        title: 'SBOM Generation Capability',
        question: 'Genereert uw build systeem <strong>automatisch een CycloneDX of SPDX SBOM</strong> bij elke productie-build?',
        tip: 'Yocto: <code>INHERIT += "create-spdx"</code> | Buildroot: manual via <code>legal-info</code>',
        article: 'CRA Art. 14.1 - SBOM verplicht vanaf 11 sept 2026, machine-readable format',
        risk: 8
    },
    {
        id: 2,
        icon: '🔍',
        title: 'CPE 2.3 Component Identification',
        question: 'Hebben <strong>alle packages in uw rootfs</strong> een CPE 2.3 identifier (cpe:/a:vendor:product:version)?',
        tip: 'Check SBOM: elk component moet CPE hebben voor CVE matching tegen NVD',
        article: 'CRA Art. 14.1 - Zonder CPE geen geautomatiseerde vulnerability tracking',
        risk: 7
    },
    {
        id: 3,
        icon: '🔧',
        title: 'Out-of-Tree Kernel Modules',
        question: 'Gebruikt u <strong>proprietary kernel modules (.ko)</strong> zonder gedocumenteerde upstream source?',
        tip: 'Elke <code>insmod</code> in init scripts moet traceerbaar zijn naar SRC_URI',
        article: 'Annex I.II.1 - Alle kernel extensions vereisen provenance tracking',
        risk: 6
    },
    {
        id: 4,
        icon: '🐛',
        title: 'Known CVE Exposure',
        question: 'Weet u <strong>hoeveel actieve CVEs</strong> er in uw huidige Linux distributie zitten?',
        tip: 'Run CVE scan tegen NVD database - vaak 50+ onbekende vulnerabilities',
        article: 'Annex I.I.1 - Geen bekende exploiteerbare kwetsbaarheden toegestaan',
        risk: 7
    },
    {
        id: 5,
        icon: '⏰',
        title: 'Vulnerability Response Time',
        question: 'Heeft u een <strong>gedocumenteerd proces</strong> voor CVE patches binnen 24 uur na disclosure?',
        tip: 'CRA vereist automated monitoring + emergency patch capability',
        article: 'CRA Art. 11.1 - Vulnerabilities gemeld binnen 24u aan ENISA',
        risk: 6
    },
    {
        id: 6,
        icon: '📡',
        title: 'CSAF 2.0 Vulnerability Endpoint',
        question: 'Publiceert u <strong>machine-readable security advisories</strong> op /.well-known/csaf/?',
        tip: 'CSAF 2.0 is EU standaard - klanten gaan dit geautomatiseerd scrapen',
        article: 'CRA Art. 11.2 - Geautomatiseerde vulnerability disclosure verplicht',
        risk: 5
    },
    {
        id: 7,
        icon: '🕐',
        title: 'Kernel LTS Support Coverage',
        question: 'Is uw Linux kernel versie <strong>EOL binnen 5 jaar</strong> na product launch?',
        tip: '<code>uname -r</code> - Check kernel.org: 4.x/5.4 nadert EOL, 6.6 LTS loopt tot 2029',
        article: 'CRA Art. 10.4 - Security updates gedurende volledige product lifetime',
        risk: 7
    },
    {
        id: 8,
        icon: '⚡',
        title: 'Secure Boot Implementation',
        question: 'Verifieert uw bootloader <strong>kernel signatures</strong> met hardware root-of-trust (TPM/TEE)?',
        tip: 'U-Boot: <code>CONFIG_FIT_SIGNATURE</code> + verified boot chain',
        article: 'Annex I.II.4 - Verified boot mandatory voor connected devices',
        risk: 6
    },
    {
        id: 9,
        icon: '🔐',
        title: 'Kernel Module Signing',
        question: 'Zijn <strong>alle kernel modules</strong> cryptografisch gesigneerd en gevalideerd at boot?',
        tip: 'Check <code>CONFIG_MODULE_SIG_FORCE=y</code> - voorkomt malicious .ko loading',
        article: 'Annex I.II.4 - Unsigned modules = attack surface voor rootkits',
        risk: 5
    },
    {
        id: 10,
        icon: '⚖️',
        title: 'GPL-3.0 Copyleft Compliance',
        question: 'Linkt uw proprietary code tegen <strong>GPLv3 libraries</strong> (readline, gmp, bash)?',
        tip: '<code>ldd /usr/bin/yourdaemon</code> - GPLv3 = source disclosure vereist',
        article: 'CRA Art. 14.2 + GPLv3 §6 - Tivoization = dubbele compliance breach',
        risk: 5
    },
    {
        id: 11,
        icon: '📄',
        title: 'License SPDX Documentation',
        question: 'Heeft u <strong>SPDX license identifiers</strong> voor alle 500+ packages in uw image?',
        tip: 'CRA vereist volledige license transparency - "Unknown" is non-compliant',
        article: 'CRA Art. 14.2 - License disclosure mandatory in SBOM',
        risk: 4
    },
    {
        id: 12,
        icon: '🌐',
        title: 'Default Credentials Exposure',
        question: 'Shipped uw device met <strong>hardcoded passwords</strong> in /etc/shadow of config files?',
        tip: 'Check init scripts: root:root, admin:admin = instant fail bij certificatie',
        article: 'Annex I.II.2 - No default credentials in production firmware',
        risk: 6
    },
    {
        id: 13,
        icon: '🔌',
        title: 'Network Service Attack Surface',
        question: 'Draait uw device <strong>onnodige daemons</strong> (telnetd, ftpd, dropbear op 0.0.0.0)?',
        tip: '<code>netstat -tulpn</code> - BusyBox telnet/FTP = red flag',
        article: 'Annex I.II.2 - Secure by default, minimal attack surface',
        risk: 5
    },
    {
        id: 14,
        icon: '🔄',
        title: 'OTA Update Cryptographic Validation',
        question: 'Valideert uw OTA updater <strong>firmware signatures</strong> voor elke update?',
        tip: 'Check update script: zonder signature validation = remote code execution risk',
        article: 'Annex I.II.5 - Secure update mechanism met rollback capability',
        risk: 7
    },
    {
        id: 15,
        icon: '🛡️',
        title: 'Memory Protection (ASLR/DEP)',
        question: 'Zijn <strong>ASLR en DEP</strong> enabled in uw kernel config en userspace?',
        tip: '<code>cat /proc/sys/kernel/randomize_va_space</code> (moet 2 zijn)',
        article: 'Annex I.II.3 - Modern memory protection mechanisms mandatory',
        risk: 4
    },
    {
        id: 16,
        icon: '📋',
        title: 'Technical Construction File',
        question: 'Heeft u een <strong>42+ pagina TCF</strong> met architecture diagrams, threat model, test results?',
        tip: 'CRA Annex VII - Dit document moet CE marking ondersteunen',
        article: 'CRA Art. 24 - Technical documentation mandatory voor markttoelating',
        risk: 6
    },
    {
        id: 17,
        icon: '🏭',
        title: 'Build Reproducibility',
        question: 'Zijn uw builds <strong>bit-for-bit reproducible</strong> vanaf dezelfde source commit?',
        tip: 'Check timestamps in binaries: non-reproducible = provenance issues',
        article: 'CRA Art. 14.1 - Reproducible builds bewijzen supply chain integrity',
        risk: 4
    },
    {
        id: 18,
        icon: '🔗',
        title: 'Supply Chain PURL Mapping',
        question: 'Heeft elk package een <strong>PURL</strong> (pkg:github/vendor/repo@version) in uw SBOM?',
        tip: 'PURL = upstream source traceability - vereist voor supply chain audits',
        article: 'CRA Art. 14.1 - Supply chain transparency via package URLs',
        risk: 4
    },
    {
        id: 19,
        icon: '⚠️',
        title: 'Critical Infrastructure Classification',
        question: 'Verkoopt u aan <strong>energie, telecom, water of transport</strong> (NIS2 sectors)?',
        tip: 'Critical infrastructure = strengere eisen (Article 6 Class I/II)',
        article: 'CRA Art. 6 - Critical products hebben aanvullende certificatie-eisen',
        risk: 5
    },
    {
        id: 20,
        icon: '📅',
        title: 'Update Lifecycle Commitment',
        question: 'Garandeert u <strong>security updates voor 5+ jaar</strong> na laatste product verkoop?',
        tip: 'Dit moet gedocumenteerd zijn in product documentation',
        article: 'CRA Art. 10.4 - Support period disclosure verplicht',
        risk: 5
    },
    {
        id: 21,
        icon: '📊',
        title: 'EU Market Conformity Declaration',
        question: 'Heeft u een <strong>EU Declaration of Conformity</strong> klaar voor CRA (vergelijkbaar met CE voor EMC)?',
        tip: 'Dit document moet refereren naar CRA Annex VII compliance',
        article: 'CRA Art. 28 - DoC verplicht voor markttoetreding vanaf dec 2027',
        risk: 4
    }
];

function initRiskWizard() {
    renderQuestion(1);
}

function renderQuestion(qNum) {
    const q = questions[qNum - 1];
    const container = document.getElementById('current-question');
    
    container.innerHTML = `
        <div class="question-card animate-fade-in">
            <div class="text-center mb-6">
                <div class="inline-flex items-center justify-center w-16 h-16 bg-orange-100 rounded-full mb-4 text-4xl">
                    ${q.icon}
                </div>
                <h3 class="text-xl font-bold text-gray-900">${q.title}</h3>
            </div>
            
            <p class="text-gray-800 text-lg mb-6 leading-relaxed text-center">${q.question}</p>
            
            <div class="bg-blue-50 border-l-4 border-blue-500 p-4 mb-8 rounded-r-lg">
                <p class="text-sm text-blue-900"><strong>💡 Technische check:</strong> ${q.tip}</p>
                <p class="text-xs text-blue-700 mt-2">📖 ${q.article}</p>
            </div>
            
            <div class="grid grid-cols-2 gap-6">
                <button type="button" onclick="answerQuestion(${qNum}, 'no', ${q.risk})" 
                    class="group answer-btn bg-gradient-to-br from-red-500 to-red-600 hover:from-red-600 hover:to-red-700 text-white font-bold py-6 px-8 rounded-xl transition-all transform hover:scale-105 shadow-lg hover:shadow-xl">
                    <div class="text-3xl mb-2">❌</div>
                    <div class="text-lg">Nee / Onzeker</div>
                    <div class="text-xs opacity-75 mt-1">+${q.risk} risk points</div>
                </button>
                <button type="button" onclick="answerQuestion(${qNum}, 'yes', 0)" 
                    class="group answer-btn bg-gradient-to-br from-green-500 to-green-600 hover:from-green-600 hover:to-green-700 text-white font-bold py-6 px-8 rounded-xl transition-all transform hover:scale-105 shadow-lg hover:shadow-xl">
                    <div class="text-3xl mb-2">✅</div>
                    <div class="text-lg">Ja, compliant</div>
                    <div class="text-xs opacity-75 mt-1">Geen risico</div>
                </button>
            </div>
        </div>
    `;
    
    // Update progress
    updateProgress();
}

function answerQuestion(qNum, answer, risk) {
    answers[qNum] = { answer, risk };
    
    if (answer === 'no') {
        riskScore += risk;
    }
    
    // Animate out
    const container = document.getElementById('current-question');
    container.style.opacity = '0';
    container.style.transform = 'translateX(-50px)';
    
    setTimeout(() => {
        currentQuestion++;
        
        if (currentQuestion <= totalQuestions) {
            container.style.opacity = '1';
            container.style.transform = 'translateX(0)';
            renderQuestion(currentQuestion);
        } else {
            showEmailForm();
        }
    }, 300);
    
    updateProgress();
}

function updateProgress() {
    // Update text
    document.getElementById('progress-text').textContent = currentQuestion > totalQuestions 
        ? 'Afgerond!' 
        : `Vraag ${currentQuestion} van ${totalQuestions}`;
    
    // Update dots - mark completed and current
    const dots = document.querySelectorAll('#progress-dots span');
    dots.forEach((dot, index) => {
        const stepNum = index + 1;
        if (stepNum < currentQuestion) {
            // Completed step
            dot.className = 'block size-2.5 rounded-full bg-orange-500 transition-all';
        } else if (stepNum === currentQuestion) {
            // Current step - larger with glow
            dot.className = 'block size-3 rounded-full bg-orange-500 ring-4 ring-orange-200 transition-all';
        } else {
            // Upcoming step
            dot.className = 'block size-2.5 rounded-full bg-gray-200 transition-all';
        }
    });
    
    // Update risk score preview
    const scoreEl = document.getElementById('risk-score-preview');
    scoreEl.textContent = `Risk: ${riskScore}/100`;
    
    if (riskScore >= 70) {
        scoreEl.className = 'text-sm font-bold text-red-600';
    } else if (riskScore >= 40) {
        scoreEl.className = 'text-sm font-bold text-orange-600';
    } else {
        scoreEl.className = 'text-sm font-bold text-green-600';
    }
}

function showEmailForm() {
    const container = document.getElementById('current-question');
    container.style.opacity = '1';
    container.style.transform = 'translateX(0)';
    
    const riskLevel = riskScore >= 70 ? 'HOOG' : (riskScore >= 40 ? 'MIDDEN' : 'LAAG');
    const emoji = riskScore >= 70 ? '⚠️' : (riskScore >= 40 ? '⚡' : '✅');
    const bgColor = riskScore >= 70 ? 'bg-red-100' : (riskScore >= 40 ? 'bg-orange-100' : 'bg-green-100');
    const textColor = riskScore >= 70 ? 'text-red-600' : (riskScore >= 40 ? 'text-orange-600' : 'text-green-600');
    const borderColor = riskScore >= 70 ? 'border-red-200' : (riskScore >= 40 ? 'border-orange-200' : 'border-green-200');
    const failedCount = Object.keys(answers).filter(k => answers[k].answer === 'no').length;
    
    container.innerHTML = `
        <div class="text-center mb-10">
            <div class="inline-flex items-center justify-center w-20 h-20 ${bgColor} rounded-full mb-4">
                <span class="text-5xl">${emoji}</span>
            </div>
            <div class="text-7xl font-bold ${textColor} mb-3">${riskScore}<span class="text-4xl">/100</span></div>
            <div class="text-2xl font-bold ${textColor} mb-2">${riskLevel} RISICO</div>
            <p class="text-gray-600 text-sm">Gebaseerd op ${failedCount} compliance gap${failedCount !== 1 ? 's' : ''}</p>
        </div>
        
        <div class="bg-white border-2 ${borderColor} rounded-2xl p-8 mb-6 shadow-sm">
            <h4 class="text-lg font-bold text-gray-900 mb-2 text-center">Ontvang uw 42-pagina Technical Report</h4>
            <p class="text-sm text-gray-600 mb-6 text-center">Compliance gaps + remediation roadmap + CRA article mapping</p>
            
            <div class="space-y-3">
                <input type="email" id="email-input" required 
                    class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500 text-base" 
                    placeholder="zakelijk.email@uwbedrijf.be">
                <input type="text" id="company-input" 
                    class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 text-base" 
                    placeholder="Bedrijfsnaam (optioneel)">
                    
                <select id="company-size-input" class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 text-base text-gray-700">
                    <option value="">Bedrijfsgrootte</option>
                    <option value="1-10">1-10 werknemers</option>
                    <option value="11-50">11-50 werknemers</option>
                    <option value="51-250" selected>51-250 werknemers</option>
                    <option value="250+">250+ werknemers</option>
                </select>
            </div>
            
            <button type="button" onclick="submitAssessment()" 
                class="w-full mt-6 bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white font-bold py-4 px-8 rounded-xl transition-all transform hover:scale-[1.02] shadow-lg text-lg">
                📧 Email Mijn Risk Report
            </button>
            
            <p class="text-xs text-gray-500 mt-4 text-center">
                ✓ Geen credit card vereist · ✓ Direct in uw inbox · ✓ 100% confidentieel
            </p>
        </div>
    `;
    
    // Mark all dots as completed
    const dots = document.querySelectorAll('#progress-dots span');
    dots.forEach(dot => {
        dot.className = 'block size-2.5 rounded-full bg-orange-500 transition-all';
    });
}

function submitAssessment() {
    const email = document.getElementById('email-input').value;
    const company = document.getElementById('company-input').value;
    const companySize = document.getElementById('company-size-input').value;
    
    if (!email || !email.includes('@')) {
        alert('Vul een geldig e-mailadres in');
        return;
    }
    
    // Set hidden form values
    document.getElementById('final-email').value = email;
    document.getElementById('final-company').value = company;
    document.getElementById('final-company-size').value = companySize || '51-250';
    
    // Map answers to hidden inputs (not checkboxes anymore)
    document.querySelector('input[name="sells_to_infrastructure"]').value = answers[1]?.answer === 'no' ? 'true' : 'false';
    document.querySelector('input[name="uses_open_source"]').value = answers[2]?.answer === 'no' ? 'true' : 'false';
    document.querySelector('input[name="has_sbom"]').value = answers[3]?.answer === 'no' ? 'true' : 'false';
    document.querySelector('input[name="has_vuln_process"]').value = answers[4]?.answer === 'no' ? 'true' : 'false';
    document.querySelector('input[name="products_in_eu"]').value = answers[5]?.answer === 'no' ? 'true' : 'false';
    document.querySelector('input[name="kernel_eol"]').value = answers[6]?.answer === 'no' ? 'true' : 'false';
    
    // Submit via HTMX
    document.getElementById('risk-form').requestSubmit();
}

// Initialize on modal open
document.addEventListener('DOMContentLoaded', function() {
    const modal = document.getElementById('risk-assessment-modal');
    modal.addEventListener('click', function(e) {
        if (e.target === modal) {
            // Reset on close
            currentQuestion = 1;
            riskScore = 0;
            answers = {};
            initRiskWizard();
        }
    });
});
