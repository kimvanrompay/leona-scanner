// Interactive CRA Risk Assessment Wizard
let currentQuestion = 1;
let totalQuestions = 6;
let riskScore = 0;
let answers = {};

const questions = [
    {
        id: 1,
        icon: '📦',
        title: 'CPE 2.3 Binary Traceability',
        question: 'Kunnen <strong>alle binaries</strong> in uw rootfs getraceerd worden naar source packages met CPE 2.3 identifiers?',
        tip: 'Run <code>opkg list-installed</code> - elk package moet een CPE hebben',
        article: 'CRA Art. 14.1 - Elk artefact moet volledig traceerbaar zijn via SBOM',
        risk: 30
    },
    {
        id: 2,
        icon: '🔧',
        title: 'Out-of-Tree Kernel Module Provenance',
        question: 'Gebruikt u <strong>proprietary kernel modules</strong> zonder gedocumenteerde SRC_URI in uw .bbappend files?',
        tip: 'Check je Yocto layers: elke <code>.ko</code> module moet een upstream source URI hebben',
        article: 'Annex I.II.1 - Provenance tracking vereist voor alle kernel extensions',
        risk: 25
    },
    {
        id: 3,
        icon: '🔐',
        title: 'CSAF 2.0 Vulnerability Disclosure',
        question: 'Heeft u een <strong>machine-readable CVE endpoint</strong> op <code>/.well-known/csaf/provider-metadata.json</code>?',
        tip: 'CRA vereist geautomatiseerde vulnerability disclosure binnen 24u',
        article: 'CRA Art. 11.2 - CSAF 2.0 is de EU standaard voor CVE disclosure',
        risk: 20
    },
    {
        id: 4,
        icon: '⚡',
        title: 'Cryptographic Boot Chain',
        question: 'Verifieert uw bootloader (U-Boot/GRUB) <strong>kernel signatures</strong> met hardware root-of-trust?',
        tip: 'Check <code>CONFIG_MODULE_SIG_FORCE=y</code> + Secure Boot in firmware',
        article: 'Annex I.II.4 - Verified boot chain verplicht voor critical infra',
        risk: 15
    },
    {
        id: 5,
        icon: '⚖️',
        title: 'GPL-3.0 Tivoization Risk',
        question: 'Linkt uw proprietary userspace code tegen <strong>GPLv3 libraries</strong> zonder source disclosure?',
        tip: 'Run <code>ldd /usr/bin/yourdaemon</code> - check voor libreadline, libgmp',
        article: 'CRA Art. 14.2 + GPLv3 §6 - Tivoization = dubbele compliance breach',
        risk: 8
    },
    {
        id: 6,
        icon: '🕐',
        title: 'Kernel EOL Lifecycle',
        question: 'Is uw Linux kernel versie <strong>EOL binnen 5 jaar</strong> na product release?',
        tip: 'Check <code>uname -r</code> - Kernel 4.x/5.4 zonder LTS = non-compliant',
        article: 'CRA Art. 10.4 - Security updates vereist gedurende volledige product lifetime',
        risk: 2
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
    const progress = ((currentQuestion - 1) / totalQuestions) * 100;
    document.getElementById('progress-bar').style.width = progress + '%';
    document.getElementById('progress-text').textContent = currentQuestion > totalQuestions 
        ? 'Afgerond!' 
        : `Vraag ${currentQuestion} van ${totalQuestions}`;
    
    // Update risk score preview
    const scoreEl = document.getElementById('risk-score-preview');
    scoreEl.textContent = `Risk: ${riskScore}/100`;
    
    if (riskScore >= 70) {
        scoreEl.className = 'font-bold text-red-600';
    } else if (riskScore >= 40) {
        scoreEl.className = 'font-bold text-orange-600';
    } else {
        scoreEl.className = 'font-bold text-green-600';
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
        <div class="text-center mb-8">
            <div class="inline-flex items-center justify-center w-24 h-24 ${bgColor} rounded-full mb-6">
                <span class="text-6xl">${emoji}</span>
            </div>
            <h3 class="text-3xl font-bold text-gray-900 mb-2">Uw CRA Risk Score</h3>
            <div class="text-6xl font-bold ${textColor} mb-2">${riskScore}/100</div>
            <div class="text-xl font-semibold ${textColor}">${riskLevel} RISICO</div>
        </div>
        
        <div class="bg-gray-50 border-2 ${borderColor} rounded-xl p-6 mb-6">
            <p class="text-center text-gray-700 mb-4">
                <strong>Ontvang uw volledige risico-analyse + remediation roadmap</strong>
            </p>
            <div class="space-y-4">
                <input type="email" id="email-input" required 
                    class="w-full px-4 py-3 border-2 border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent text-lg" 
                    placeholder="uw.email@bedrijf.be">
                <input type="text" id="company-input" 
                    class="w-full px-4 py-3 border-2 border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 text-lg" 
                    placeholder="Bedrijfsnaam (optioneel)">
            </div>
        </div>
        
        <button type="button" onclick="submitAssessment()" 
            class="w-full bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white font-bold py-4 px-8 rounded-xl transition-all transform hover:scale-105 shadow-lg text-lg">
            📧 Email Mijn Risk Report
        </button>
        
        <p class="text-xs text-gray-500 mt-4 text-center">
            We sturen een 42-pagina technisch rapport met specifieke fixes voor uw ${failedCount} compliance gaps
        </p>
    `;
    
    document.getElementById('progress-bar').style.width = '100%';
}

function submitAssessment() {
    const email = document.getElementById('email-input').value;
    const company = document.getElementById('company-input').value;
    
    if (!email || !email.includes('@')) {
        alert('Vul een geldig e-mailadres in');
        return;
    }
    
    // Set hidden form values
    document.getElementById('final-email').value = email;
    document.getElementById('final-company').value = company;
    
    // Map answers to form checkboxes
    document.querySelector('input[name="sells_to_infrastructure"]').checked = answers[1]?.answer === 'no';
    document.querySelector('input[name="uses_open_source"]').checked = answers[2]?.answer === 'no';
    document.querySelector('input[name="has_sbom"]').checked = answers[3]?.answer === 'no';
    document.querySelector('input[name="has_vuln_process"]').checked = answers[4]?.answer === 'no';
    document.querySelector('input[name="products_in_eu"]').checked = answers[5]?.answer === 'no';
    
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
