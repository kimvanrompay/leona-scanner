// Risk Assessment Wizard Controller (Stimulus)
import { Controller } from "https://cdn.skypack.dev/@hotwired/stimulus@3.2.2";

export default class extends Controller {
    static targets = ["progressBar", "progressText", "riskScore", "question"];
    static values = {
        currentQuestion: { type: Number, default: 1 },
        totalQuestions: { type: Number, default: 6 },
        riskScore: { type: Number, default: 0 }
    };

    connect() {
        console.log("🧙 Wizard controller connected");
        this.answers = {};
        this.updateProgress();
    }

    // Answer a question
    answer(event) {
        const button = event.currentTarget;
        const questionNum = parseInt(button.dataset.question);
        const answerValue = button.dataset.answer;
        const riskPoints = parseInt(button.dataset.risk || 0);

        // Store answer
        this.answers[questionNum] = {
            answer: answerValue,
            risk: riskPoints
        };

        // Update risk score
        if (answerValue === 'no') {
            this.riskScoreValue += riskPoints;
        }

        // Animate out current question
        this.questionTarget.style.opacity = '0';
        this.questionTarget.style.transform = 'translateX(-50px)';

        setTimeout(() => {
            this.currentQuestionValue++;
            
            if (this.currentQuestionValue <= this.totalQuestionsValue) {
                // Load next question via Turbo Frame
                const frame = document.querySelector('turbo-frame#wizard-question');
                frame.src = `/wizard/question/${this.currentQuestionValue}`;
                
                this.questionTarget.style.opacity = '1';
                this.questionTarget.style.transform = 'translateX(0)';
            } else {
                // Load email form
                const frame = document.querySelector('turbo-frame#wizard-question');
                frame.src = `/wizard/result?score=${this.riskScoreValue}&answers=${JSON.stringify(this.answers)}`;
            }
            
            this.updateProgress();
        }, 300);
    }

    updateProgress() {
        const progress = ((this.currentQuestionValue - 1) / this.totalQuestionsValue) * 100;
        
        if (this.hasProgressBarTarget) {
            this.progressBarTarget.style.width = `${progress}%`;
        }
        
        if (this.hasProgressTextTarget) {
            this.progressTextTarget.textContent = this.currentQuestionValue > this.totalQuestionsValue 
                ? 'Afgerond!' 
                : `Vraag ${this.currentQuestionValue} van ${this.totalQuestionsValue}`;
        }
        
        if (this.hasRiskScoreTarget) {
            this.riskScoreTarget.textContent = `Risk: ${this.riskScoreValue}/100`;
            
            // Update color
            if (this.riskScoreValue >= 70) {
                this.riskScoreTarget.className = 'font-bold text-red-600';
            } else if (this.riskScoreValue >= 40) {
                this.riskScoreTarget.className = 'font-bold text-orange-600';
            } else {
                this.riskScoreTarget.className = 'font-bold text-green-600';
            }
        }
    }

    submitAssessment(event) {
        event.preventDefault();
        
        const email = document.getElementById('email-input').value;
        const company = document.getElementById('company-input').value;
        
        if (!email || !email.includes('@')) {
            alert('Vul een geldig e-mailadres in');
            return;
        }
        
        // Submit via Turbo (automatically handled by form)
        const form = this.element.querySelector('form');
        if (form) {
            // Set answers data
            const answersInput = document.createElement('input');
            answersInput.type = 'hidden';
            answersInput.name = 'answers';
            answersInput.value = JSON.stringify(this.answers);
            form.appendChild(answersInput);
            
            form.requestSubmit();
        }
    }
}

// Register controller
if (window.Stimulus) {
    window.Stimulus.register("wizard", WizardController);
}
