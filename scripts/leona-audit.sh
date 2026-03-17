#!/bin/sh
# LEONA | Hardware Attestation Script v1.1
# Compliance: CRA Annex I / NIS2 / CER
# 
# Usage: Run this script on your target embedded Linux device
#        Upload the generated JSON file to your LEONA dashboard
#
# Requirements: POSIX-compliant shell, basic Linux utilities

set -e

echo "========================================="
echo "LEONA Hardware Attestation"
echo "Version 1.1 | March 2026"
echo "========================================="
echo ""

OUTPUT_FILE="leona_attestation_$(hostname)_$(date +%s).json"

echo "Collecting hardware security data..."
echo ""

# Start JSON output
echo "{" > "$OUTPUT_FILE"
echo "  \"audit_version\": \"1.1\"," >> "$OUTPUT_FILE"
echo "  \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"," >> "$OUTPUT_FILE"
echo "  \"hostname\": \"$(hostname)\"," >> "$OUTPUT_FILE"
echo "  \"data\": {" >> "$OUTPUT_FILE"

# 1. Secure Boot Status (CRA Art. 3.1)
echo "[1/8] Checking Secure Boot status..."
if [ -d /sys/firmware/efi/efivars ]; then
    SECURE_BOOT="detected"
else
    SECURE_BOOT="not_found"
fi
echo "    \"secure_boot\": \"$SECURE_BOOT\"," >> "$OUTPUT_FILE"

# 2. Kernel Version & LTS Status
echo "[2/8] Detecting kernel version..."
KERNEL_VERSION=$(uname -r)
echo "    \"kernel_version\": \"$KERNEL_VERSION\"," >> "$OUTPUT_FILE"

# 3. Kernel Hardening Check (CRA Art. 3.2)
echo "[3/8] Auditing kernel hardening features..."
if [ -f /proc/config.gz ]; then
    STACK_PROTECTOR=$(zgrep CONFIG_SCHED_STACK_END_CHECK /proc/config.gz 2>/dev/null | cut -d'=' -f2 || echo "disabled")
    KASLR=$(zgrep CONFIG_RANDOMIZE_BASE /proc/config.gz 2>/dev/null | cut -d'=' -f2 || echo "disabled")
else
    STACK_PROTECTOR="unknown"
    KASLR="unknown"
fi
echo "    \"kernel_hardening\": {" >> "$OUTPUT_FILE"
echo "      \"stack_protector\": \"$STACK_PROTECTOR\"," >> "$OUTPUT_FILE"
echo "      \"kaslr\": \"$KASLR\"" >> "$OUTPUT_FILE"
echo "    }," >> "$OUTPUT_FILE"

# 4. Read-Only RootFS check (CRA Art. 3.2 - Tampering Protection)
echo "[4/8] Checking filesystem integrity..."
if mount | grep ' / ' | grep -q '(ro'; then
    READONLY_FS="true"
else
    READONLY_FS="false"
fi
echo "    \"readonly_rootfs\": $READONLY_FS," >> "$OUTPUT_FILE"

# 5. Entropy / RNG Check (NIS2 Cryptography)
echo "[5/8] Measuring entropy availability..."
if [ -f /proc/sys/kernel/random/entropy_avail ]; then
    ENTROPY=$(cat /proc/sys/kernel/random/entropy_avail)
else
    ENTROPY=0
fi
echo "    \"entropy_avail\": $ENTROPY," >> "$OUTPUT_FILE"

# 6. Open Ports Audit (Physical Interface Security)
echo "[6/8] Scanning open network ports..."
if command -v netstat >/dev/null 2>&1; then
    OPEN_PORTS=$(netstat -tuln 2>/dev/null | grep LISTEN | awk '{print $4}' | sed 's/.*://' | sort -u | tr '\n' ',' | sed 's/,$//')
elif command -v ss >/dev/null 2>&1; then
    OPEN_PORTS=$(ss -tuln 2>/dev/null | grep LISTEN | awk '{print $5}' | sed 's/.*://' | sort -u | tr '\n' ',' | sed 's/,$//')
else
    OPEN_PORTS="unknown"
fi
echo "    \"open_ports\": \"$OPEN_PORTS\"," >> "$OUTPUT_FILE"

# 7. TPM Status (Hardware Root of Trust)
echo "[7/8] Detecting TPM module..."
if [ -d /sys/class/tpm/tpm0 ]; then
    TPM_STATUS="detected"
elif [ -c /dev/tpm0 ]; then
    TPM_STATUS="detected"
else
    TPM_STATUS="not_found"
fi
echo "    \"tpm_module\": \"$TPM_STATUS\"," >> "$OUTPUT_FILE"

# 8. SELinux / AppArmor Status
echo "[8/8] Checking mandatory access control..."
if command -v getenforce >/dev/null 2>&1; then
    MAC_STATUS="selinux_$(getenforce 2>/dev/null | tr '[:upper:]' '[:lower:]')"
elif [ -d /sys/kernel/security/apparmor ]; then
    MAC_STATUS="apparmor_enabled"
else
    MAC_STATUS="none"
fi
echo "    \"mandatory_access_control\": \"$MAC_STATUS\"" >> "$OUTPUT_FILE"

# Close JSON
echo "  }" >> "$OUTPUT_FILE"
echo "}" >> "$OUTPUT_FILE"

echo ""
echo "========================================="
echo "✓ Attestation Complete"
echo "========================================="
echo ""
echo "Report generated: $OUTPUT_FILE"
echo ""
echo "Next steps:"
echo "1. Review the attestation data in $OUTPUT_FILE"
echo "2. Upload this file to your LEONA dashboard"
echo "3. Complete the hardware validation process"
echo ""
echo "For support: support@leona-cravit.be"
echo "========================================="
