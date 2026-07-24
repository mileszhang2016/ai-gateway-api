import json
import os
import sys
from datetime import datetime

# Read test output
json_path = os.path.join(os.path.dirname(__file__), 'test-reports', 'ai_route_test_output.json')
with open(json_path, 'r', encoding='utf-8-sig') as f:
    lines = [line.strip() for line in f if line.strip()]

# Parse JSON lines
events = []
for line in lines:
    try:
        events.append(json.loads(line))
    except:
        pass

# Extract test results
results = {}  # package -> {tests: [{name, status, output, elapsed}]}
packages = {}  # package -> {status, elapsed}

for evt in events:
    action = evt.get('Action', '')
    pkg = evt.get('Package', '')
    test = evt.get('Test', '')
    
    if action == 'run' and test:
        if pkg not in results:
            results[pkg] = {}
        results[pkg][test] = {
            'name': test,
            'status': 'running',
            'output': '',
            'elapsed': 0
        }
    elif action == 'output' and test:
        if pkg in results and test in results[pkg]:
            results[pkg][test]['output'] += evt.get('Output', '')
    elif action == 'pass' and test:
        if pkg in results and test in results[pkg]:
            results[pkg][test]['status'] = 'pass'
            results[pkg][test]['elapsed'] = evt.get('Elapsed', 0)
    elif action == 'fail' and test:
        if pkg in results and test in results[pkg]:
            results[pkg][test]['status'] = 'fail'
            results[pkg][test]['elapsed'] = evt.get('Elapsed', 0)
    elif action == 'pass' and not test:
        packages[pkg] = {'status': 'pass', 'elapsed': evt.get('Elapsed', 0)}
    elif action == 'fail' and not test:
        packages[pkg] = {'status': 'fail', 'elapsed': evt.get('Elapsed', 0)}

# Generate report
report_dir = os.path.join(os.path.dirname(__file__), 'test-reports')
timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
report_subdir = os.path.join(report_dir, f'ai_route_{timestamp}')
os.makedirs(report_subdir, exist_ok=True)

# Calculate totals
total_tests = 0
passed_tests = 0
failed_tests = 0
for pkg, tests in results.items():
    for name, info in tests.items():
        total_tests += 1
        if info['status'] == 'pass':
            passed_tests += 1
        elif info['status'] == 'fail':
            failed_tests += 1

# Generate summary
summary = []
summary.append(f'# AI 路由规则模块 - 测试报告\n')
summary.append(f'**生成时间**：{datetime.now().strftime("%Y-%m-%d %H:%M:%S")}  \n')
summary.append(f'**测试模块**：ai_route（AI 路由规则）  \n')
summary.append(f'**接口数**：2（PATCH `/ai-route-rules`、GET `/ai-route-rules`）\n\n')

summary.append(f'## 汇总统计\n\n')
summary.append(f'| 指标 | 数值 |\n')
summary.append(f'|------|------|\n')
summary.append(f'| 总用例数 | {total_tests} |\n')
summary.append(f'| 通过 | {passed_tests} |\n')
summary.append(f'| 失败 | {failed_tests} |\n')
summary.append(f'| 通过率 | {passed_tests/total_tests*100:.1f}% |\n\n' if total_tests > 0 else '| 通过率 | N/A |\n\n')

# Module details
summary.append(f'## 模块详情\n\n')

for pkg_name in sorted(results.keys()):
    tests = results[pkg_name]
    pkg_total = len(tests)
    pkg_pass = sum(1 for t in tests.values() if t['status'] == 'pass')
    pkg_fail = sum(1 for t in tests.values() if t['status'] == 'fail')
    
    short_name = pkg_name.split('/')[-1]
    summary.append(f'### {short_name}\n\n')
    summary.append(f'| 指标 | 数值 |\n')
    summary.append(f'|------|------|\n')
    summary.append(f'| 总用例数 | {pkg_total} |\n')
    summary.append(f'| 通过 | {pkg_pass} |\n')
    summary.append(f'| 失败 | {pkg_fail} |\n')
    summary.append(f'| 通过率 | {pkg_pass/pkg_total*100:.1f}% |\n\n')
    
    # Test list
    summary.append(f'| 状态 | 用例名称 | 耗时 |\n')
    summary.append(f'|------|---------|------|\n')
    for name in sorted(tests.keys()):
        info = tests[name]
        status_icon = '✅' if info['status'] == 'pass' else '❌'
        elapsed = f'{info["elapsed"]:.2f}s' if info['elapsed'] else '-'
        summary.append(f'| {status_icon} | {name} | {elapsed} |\n')
    summary.append('\n')

# Detailed test cases
summary.append(f'## 用例执行明细\n\n')

for pkg_name in sorted(results.keys()):
    tests = results[pkg_name]
    short_name = pkg_name.split('/')[-1]
    summary.append(f'### {short_name}\n\n')
    
    for name in sorted(tests.keys()):
        info = tests[name]
        status_icon = '✅ PASS' if info['status'] == 'pass' else '❌ FAIL'
        summary.append(f'#### {status_icon} - {name}\n\n')
        
        # Extract key info from output
        output = info['output']
        
        # Extract API call info
        summary.append(f'- **状态**：{info["status"]}\n')
        if info['elapsed']:
            summary.append(f'- **耗时**：{info["elapsed"]:.2f}s\n')
        
        # Extract status_code and ret_msg from log
        import re
        status_match = re.search(r'status_code\[(\d+)\]', output)
        ret_msg_match = re.search(r'ret_msg\[(.*?)\]', output)
        
        if status_match:
            summary.append(f'- **HTTP 状态码**：{status_match.group(1)}\n')
        if ret_msg_match:
            summary.append(f'- **返回消息**：{ret_msg_match.group(1)}\n')
        
        # Extract error messages from test output
        error_lines = [line.strip() for line in output.split('\n') 
                      if 'expected' in line.lower() or 'error' in line.lower() or 'fatal' in line.lower()]
        if error_lines:
            summary.append(f'\n**错误详情**：\n')
            for line in error_lines:
                summary.append(f'- {line}\n')
        
        # Extract successful assertions
        pass_lines = [line.strip() for line in output.split('\n') if '--- PASS' in line or '--- FAIL' in line]
        if pass_lines:
            summary.append(f'\n**结果**：{pass_lines[0]}\n')
        
        summary.append('\n---\n\n')

# Write report
report_path = os.path.join(report_subdir, 'ai_route_test_report.md')
with open(report_path, 'w', encoding='utf-8') as f:
    f.write(''.join(summary))

print(f'Report generated: {report_path}')
print(f'Total: {total_tests}, Pass: {passed_tests}, Fail: {failed_tests}')