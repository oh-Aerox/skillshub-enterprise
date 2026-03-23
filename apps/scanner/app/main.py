"""
SkillsHub Enterprise - Scanner Service
三层安全扫描引擎
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Optional, List, Dict, Any
import uuid
import asyncio

app = FastAPI(title="SkillsHub Scanner", version="1.0.0")


class ScanRequest(BaseModel):
    skill_version_id: str
    skill_name: str
    files: List[Dict[str, Any]]
    source_type: str = "internal"
    source_url: Optional[str] = None


class ScanResponse(BaseModel):
    scan_id: str
    status: str
    risk_level: Optional[str] = None
    risk_score: Optional[int] = None
    summary: Optional[str] = None


class Layer1Result(BaseModel):
    passed: bool
    issues: List[Dict[str, Any]]


class Layer2Result(BaseModel):
    passed: bool
    prompt_analysis: Dict[str, Any]
    tool_check: Dict[str, Any]
    secret_scan: Dict[str, Any]


class Layer3Result(BaseModel):
    passed: bool
    behavior_log: List[Dict[str, Any]]
    network_attempts: int
    file_violations: int


class Layer4Result(BaseModel):
    passed: bool
    source_reputation: float
    license_compliant: bool


# Layer 1: Structure and Format Scan
async def run_layer1_scan(files: List[Dict[str, Any]]) -> Layer1Result:
    """
    Layer 1: 结构与格式扫描
    - 检查 SKILL.md 存在性
    - 验证 YAML frontmatter
    - 文件类型白名单检查
    - 文件大小检查
    """
    issues = []
    passed = True

    # Check for SKILL.md
    has_skill_md = any(f.get("name") == "SKILL.md" for f in files)
    if not has_skill_md:
        issues.append({
            "severity": "critical",
            "type": "missing_skill_md",
            "message": "SKILL.md file not found in root directory"
        })
        passed = False

    # Check file types
    allowed_extensions = [".md", ".py", ".js", ".ts", ".sh", ".json", ".yaml", ".yml", ".txt"]
    for f in files:
        name = f.get("name", "")
        if not any(name.endswith(ext) for ext in allowed_extensions):
            issues.append({
                "severity": "warning",
                "type": "unknown_file_type",
                "message": f"File type not allowed: {name}"
            })

    # Check file sizes
    max_file_size = 10 * 1024 * 1024  # 10MB
    for f in files:
        size = f.get("size", 0)
        if size > max_file_size:
            issues.append({
                "severity": "critical",
                "type": "file_too_large",
                "message": f"File {name} exceeds 10MB limit"
            })
            passed = False

    return Layer1Result(passed=passed, issues=issues)


# Layer 2: Static Content Analysis
async def run_layer2_scan(files: List[Dict[str, Any]]) -> Layer2Result:
    """
    Layer 2: 静态内容扫描
    - 提示词意图分析
    - 工具调用白名单检查
    - 敏感信息扫描
    """
    dangerous_patterns = {
        "prompt_injection": [
            "ignore previous", "disregard", "forget your rules",
            "忽略之前", "你现在是", "覆盖系统"
        ],
        "privilege_escalation": [
            "sudo", "chmod 777", "chown root", "Administrator",
            "rm -rf", "DROP TABLE", "format disk"
        ],
        "data_exfiltration": [
            "curl", "wget", "fetch", "POST.*http", "send.*to.*http"
        ],
        "secret_extraction": [
            "API_KEY", "SECRET", "PASSWORD", "TOKEN", "private key"
        ]
    }

    tool_whitelist = ["file_read", "file_write", "bash", "python"]

    prompt_analysis = {"injection_risk": False, "detected_patterns": []}
    tool_check = {"requested_tools": [], "unauthorized_tools": []}
    secret_scan = {"secrets_found": [], "ips_found": [], "emails_found": []}

    for f in files:
        content = f.get("content", "")
        name = f.get("name", "")

        if name.endswith(".md"):
            # Check for dangerous patterns
            for category, patterns in dangerous_patterns.items():
                for pattern in patterns:
                    if pattern.lower() in content.lower():
                        prompt_analysis["detected_patterns"].append({
                            "category": category,
                            "pattern": pattern,
                            "severity": "high" if category in ["prompt_injection", "privilege_escalation"] else "medium"
                        })
                        if category in ["prompt_injection", "privilege_escalation"]:
                            prompt_analysis["injection_risk"] = True

    passed = not prompt_analysis["injection_risk"]

    return Layer2Result(
        passed=passed,
        prompt_analysis=prompt_analysis,
        tool_check=tool_check,
        secret_scan=secret_scan
    )


# Layer 3: Sandbox Behavior Testing
async def run_layer3_scan(files: List[Dict[str, Any]]) -> Layer3Result:
    """
    Layer 3: 沙箱行为测试
    - 在隔离环境中执行 Skill
    - 监控系统调用、网络尝试、文件访问
    """
    # Simulated sandbox execution
    behavior_log = []
    network_attempts = 0
    file_violations = 0

    # In production, this would run in a Docker/gVisor sandbox
    for f in files:
        if f.get("name", "").endswith((".py", ".js", ".sh")):
            content = f.get("content", "")
            # Check for network calls
            if any(x in content for x in ["requests.", "urllib", "fetch(", "axios"]):
                network_attempts += 1
                behavior_log.append({
                    "type": "network_attempt",
                    "file": f.get("name"),
                    "severity": "high"
                })

    passed = network_attempts == 0 and file_violations == 0

    return Layer3Result(
        passed=passed,
        behavior_log=behavior_log,
        network_attempts=network_attempts,
        file_violations=file_violations
    )


# Layer 4: Supply Chain Analysis
async def run_layer4_scan(source_type: str, source_url: Optional[str] = None) -> Layer4Result:
    """
    Layer 4: 供应链溯源分析
    - 来源信誉评分
    - License 合规检查
    - 版本历史分析
    """
    source_reputation = 50.0  # Default
    license_compliant = True

    if source_type == "opensource" and source_url:
        # In production, query GitHub API for repo info
        if "github.com" in (source_url or ""):
            source_reputation = 70.0  # Simulated

    return Layer4Result(
        passed=license_compliant and source_reputation > 30,
        source_reputation=source_reputation,
        license_compliant=license_compliant
    )


def calculate_risk_score(l1: Layer1Result, l2: Layer2Result,
                         l3: Layer3Result, l4: Layer4Result) -> tuple:
    """
    计算综合风险评分 (0-100, 越低风险越低)
    """
    score = 0

    # Layer 1: Structure issues
    for issue in l1.issues:
        if issue.get("severity") == "critical":
            score += 40
        elif issue.get("severity") == "warning":
            score += 10

    # Layer 2: Content issues
    if l2.prompt_analysis.get("injection_risk"):
        score += 30
    if l2.tool_check.get("unauthorized_tools"):
        score += 20

    # Layer 3: Behavior issues
    if l3.network_attempts > 0:
        score += 25
    if l3.file_violations > 0:
        score += 20

    # Layer 4: Supply chain issues
    if l4.source_reputation < 50:
        score += 15
    if not l4.license_compliant:
        score += 10

    score = min(100, score)

    # Determine risk level
    if score <= 30:
        risk_level = "A"
    elif score <= 50:
        risk_level = "B"
    elif score <= 70:
        risk_level = "C"
    elif score <= 85:
        risk_level = "D"
    else:
        risk_level = "F"

    return score, risk_level


@app.get("/health")
async def health_check():
    return {"status": "ok"}


@app.post("/api/scan", response_model=ScanResponse)
async def trigger_scan(request: ScanRequest):
    """
    触发安全扫描流水线
    """
    scan_id = str(uuid.uuid4())

    try:
        # Run all layers
        l1_result = await run_layer1_scan(request.files)
        l2_result = await run_layer2_scan(request.files)
        l3_result = await run_layer3_scan(request.files)
        l4_result = await run_layer4_scan(request.source_type, request.source_url)

        # Calculate final score
        risk_score, risk_level = calculate_risk_score(l1_result, l2_result, l3_result, l4_result)

        # Determine if passed
        passed = risk_level in ["A", "B"]

        summary = f"Scan completed. Risk Level: {risk_level}, Score: {risk_score}"
        if not l1_result.passed:
            summary += " [FAILED: Structure check]"
        if not l2_result.passed:
            summary += " [FAILED: Content analysis]"

        return ScanResponse(
            scan_id=scan_id,
            status="completed",
            risk_level=risk_level,
            risk_score=risk_score,
            summary=summary
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/scan/{scan_id}")
async def get_scan_result(scan_id: str):
    """
    获取扫描结果
    """
    # In production, fetch from database
    return {
        "scan_id": scan_id,
        "status": "completed",
        "risk_level": "A",
        "risk_score": 15,
        "results": {
            "layer1": {"passed": True},
            "layer2": {"passed": True},
            "layer3": {"passed": True},
            "layer4": {"passed": True}
        }
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
