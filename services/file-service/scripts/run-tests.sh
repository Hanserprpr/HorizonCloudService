#!/bin/bash

# run-tests.sh - 文件服务测试运行脚本
# 提供完整的测试运行和报告功能

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置
VERBOSE=${VERBOSE:-false}
COVERAGE=${COVERAGE:-true}
BENCHMARK=${BENCHMARK:-false}
INTEGRATION=${INTEGRATION:-false}
STRESS=${STRESS:-false}
OUTPUT_DIR=${OUTPUT_DIR:-"./test-results"}
GO_TEST_FLAGS=${GO_TEST_FLAGS:-""}

# 函数定义
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
文件服务测试运行脚本

用法: $0 [选项]

选项:
    -h, --help          显示帮助信息
    -v, --verbose       详细输出模式
    -c, --coverage      启用代码覆盖率 (默认: true)
    -b, --benchmark     运行基准测试
    -i, --integration   运行集成测试
    -s, --stress        运行压力测试
    -o, --output DIR    输出目录 (默认: ./test-results)
    --unit-only         只运行单元测试
    --no-coverage       禁用代码覆盖率
    --clean             运行前清理测试缓存

示例:
    $0                          # 运行标准测试套件
    $0 -v -c                   # 详细模式 + 覆盖率
    $0 -b                      # 运行基准测试
    $0 -i                      # 运行集成测试
    $0 -s                      # 运行压力测试
    $0 --unit-only             # 只运行单元测试
    $0 -v -c -b -i             # 运行所有测试

环境变量:
    GO_TEST_FLAGS      额外的 go test 参数
    TEST_TIMEOUT       测试超时时间 (默认: 30m)
    PARALLEL_JOBS      并行任务数 (默认: CPU核心数)
EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -c|--coverage)
                COVERAGE=true
                shift
                ;;
            --no-coverage)
                COVERAGE=false
                shift
                ;;
            -b|--benchmark)
                BENCHMARK=true
                shift
                ;;
            -i|--integration)
                INTEGRATION=true
                shift
                ;;
            -s|--stress)
                STRESS=true
                shift
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            --unit-only)
                INTEGRATION=false
                BENCHMARK=false
                STRESS=false
                shift
                ;;
            --clean)
                CLEAN_CACHE=true
                shift
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查环境
check_environment() {
    log_info "检查环境..."
    
    # 设置Go路径
    GO_CMD="/usr/local/go/bin/go"
    if [ ! -f "$GO_CMD" ]; then
        # 如果直接路径不存在，尝试使用PATH中的go
        if command -v go &> /dev/null; then
            GO_CMD="go"
        else
            log_error "Go 未安装或不在 PATH 中"
            exit 1
        fi
    fi
    
    GO_VERSION=$($GO_CMD version | cut -d' ' -f3)
    log_info "Go 版本: $GO_VERSION"
    
    # 检查依赖
    if [[ ! -f "go.mod" ]]; then
        log_error "go.mod 文件不存在，请确保在正确的目录中运行"
        exit 1
    fi
    
    # 创建输出目录
    mkdir -p "$OUTPUT_DIR"
    
    # 清理缓存（如果需要）
    if [[ "${CLEAN_CACHE:-false}" == "true" ]]; then
        log_info "清理测试缓存..."
        $GO_CMD clean -testcache
    fi
    
    log_success "环境检查完成"
}

# 运行单元测试
run_unit_tests() {
    log_info "运行单元测试..."
    
    local test_flags=""
    local coverage_flags=""
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_flags+=" -v"
    fi
    
    if [[ "$COVERAGE" == "true" ]]; then
        coverage_flags=" -coverprofile=$OUTPUT_DIR/coverage.out -covermode=atomic"
        test_flags+="$coverage_flags"
    fi
    
    # 设置超时时间
    local timeout="${TEST_TIMEOUT:-30m}"
    test_flags+=" -timeout=$timeout"
    
    # 并行执行
    local parallel="${PARALLEL_JOBS:-$(nproc)}"
    test_flags+=" -parallel=$parallel"
    
    # 运行测试
    local packages=$($GO_CMD list ./... | grep -v test | grep -v cmd)
    
    if $GO_CMD test $test_flags $GO_TEST_FLAGS $packages; then
        log_success "单元测试通过"
    else
        log_error "单元测试失败"
        return 1
    fi
    
    # 生成覆盖率报告
    if [[ "$COVERAGE" == "true" && -f "$OUTPUT_DIR/coverage.out" ]]; then
        generate_coverage_report
    fi
}

# 生成覆盖率报告
generate_coverage_report() {
    log_info "生成覆盖率报告..."
    
    # 生成HTML报告
    $GO_CMD tool cover -html="$OUTPUT_DIR/coverage.out" -o "$OUTPUT_DIR/coverage.html"
    
    # 生成总体覆盖率
    local coverage_percent=$($GO_CMD tool cover -func="$OUTPUT_DIR/coverage.out" | tail -1 | awk '{print $3}')
    log_info "代码覆盖率: $coverage_percent"
    
    # 生成详细的覆盖率报告
    $GO_CMD tool cover -func="$OUTPUT_DIR/coverage.out" > "$OUTPUT_DIR/coverage.txt"
    
    log_success "覆盖率报告生成完成"
    log_info "HTML报告: $OUTPUT_DIR/coverage.html"
    log_info "文本报告: $OUTPUT_DIR/coverage.txt"
}

# 运行集成测试
run_integration_tests() {
    if [[ "$INTEGRATION" != "true" ]]; then
        return 0
    fi
    
    log_info "运行集成测试..."
    
    # 设置集成测试环境变量
    export INTEGRATION_TESTS=1
    
    local test_flags="-v"
    if [[ "$VERBOSE" == "true" ]]; then
        test_flags+=" -test.v"
    fi
    
    # 运行集成测试
    if $GO_CMD test $test_flags ./test -run Integration; then
        log_success "集成测试通过"
    else
        log_error "集成测试失败"
        return 1
    fi
}

# 运行基准测试
run_benchmark_tests() {
    if [[ "$BENCHMARK" != "true" ]]; then
        return 0
    fi
    
    log_info "运行基准测试..."
    
    # 设置基准测试环境变量
    export BENCHMARK_ALL=1
    
    # 运行基准测试
    local bench_flags="-bench=. -benchmem"
    if [[ "$VERBOSE" == "true" ]]; then
        bench_flags+=" -v"
    fi
    
    local bench_output="$OUTPUT_DIR/benchmark.txt"
    
    if $GO_CMD test $bench_flags ./test -run='^$' > "$bench_output" 2>&1; then
        log_success "基准测试完成"
        log_info "基准测试结果: $bench_output"
        
        # 显示基准测试摘要
        if [[ "$VERBOSE" == "true" ]]; then
            echo "基准测试摘要:"
            grep "Benchmark" "$bench_output" | head -10
        fi
    else
        log_error "基准测试失败"
        cat "$bench_output"
        return 1
    fi
}

# 运行压力测试
run_stress_tests() {
    if [[ "$STRESS" != "true" ]]; then
        return 0
    fi
    
    log_info "运行压力测试..."
    log_warning "压力测试可能需要大量时间和资源"
    
    # 设置压力测试环境变量
    export STRESS_TESTS=1
    
    # 运行压力测试
    local stress_flags="-bench=Stress -benchtime=30s"
    if [[ "$VERBOSE" == "true" ]]; then
        stress_flags+=" -v"
    fi
    
    local stress_output="$OUTPUT_DIR/stress.txt"
    
    if $GO_CMD test $stress_flags ./test -run='^$' > "$stress_output" 2>&1; then
        log_success "压力测试完成"
        log_info "压力测试结果: $stress_output"
    else
        log_error "压力测试失败"
        cat "$stress_output"
        return 1
    fi
}

# 运行代码质量检查
run_quality_checks() {
    log_info "运行代码质量检查..."
    
    # 运行 go vet
    if $GO_CMD vet ./...; then
        log_success "go vet 检查通过"
    else
        log_error "go vet 检查失败"
        return 1
    fi
    
    # 检查是否安装了 golangci-lint
    if command -v golangci-lint &> /dev/null; then
        log_info "运行 golangci-lint..."
        if golangci-lint run --out-format=colored-line-number > "$OUTPUT_DIR/lint.txt" 2>&1; then
            log_success "golangci-lint 检查通过"
        else
            log_warning "golangci-lint 发现问题，查看 $OUTPUT_DIR/lint.txt"
            if [[ "$VERBOSE" == "true" ]]; then
                cat "$OUTPUT_DIR/lint.txt"
            fi
        fi
    else
        log_warning "golangci-lint 未安装，跳过 linting 检查"
    fi
}

# 生成测试报告
generate_test_report() {
    log_info "生成测试报告..."
    
    local report_file="$OUTPUT_DIR/test-report.md"
    
    cat > "$report_file" << EOF
# 文件服务测试报告

**生成时间**: $(date)
**Go版本**: $($GO_CMD version | cut -d' ' -f3)

## 测试配置

- 详细输出: $VERBOSE
- 代码覆盖率: $COVERAGE
- 基准测试: $BENCHMARK
- 集成测试: $INTEGRATION
- 压力测试: $STRESS

## 测试结果

### 单元测试
EOF

    if [[ -f "$OUTPUT_DIR/coverage.out" ]]; then
        local coverage_percent=$($GO_CMD tool cover -func="$OUTPUT_DIR/coverage.out" | tail -1 | awk '{print $3}')
        echo "- 代码覆盖率: $coverage_percent" >> "$report_file"
        echo "- 覆盖率报告: [coverage.html](coverage.html)" >> "$report_file"
    fi

    if [[ "$INTEGRATION" == "true" ]]; then
        echo -e "\n### 集成测试\n- 状态: 完成" >> "$report_file"
    fi

    if [[ "$BENCHMARK" == "true" && -f "$OUTPUT_DIR/benchmark.txt" ]]; then
        echo -e "\n### 基准测试\n- 结果: [benchmark.txt](benchmark.txt)" >> "$report_file"
    fi

    if [[ "$STRESS" == "true" && -f "$OUTPUT_DIR/stress.txt" ]]; then
        echo -e "\n### 压力测试\n- 结果: [stress.txt](stress.txt)" >> "$report_file"
    fi

    echo -e "\n### 代码质量检查" >> "$report_file"
    if [[ -f "$OUTPUT_DIR/lint.txt" ]]; then
        echo "- Lint报告: [lint.txt](lint.txt)" >> "$report_file"
    fi

    log_success "测试报告生成完成: $report_file"
}

# 清理函数
cleanup() {
    # 清理临时文件和环境变量
    unset INTEGRATION_TESTS
    unset BENCHMARK_ALL
    unset STRESS_TESTS
}

# 主函数
main() {
    # 设置清理陷阱
    trap cleanup EXIT
    
    log_info "文件服务测试套件开始运行..."
    
    # 解析参数
    parse_args "$@"
    
    # 检查环境
    check_environment
    
    # 记录开始时间
    local start_time=$(date +%s)
    
    # 运行测试套件
    local failed_tests=()
    
    # 代码质量检查
    if ! run_quality_checks; then
        failed_tests+=("质量检查")
    fi
    
    # 单元测试
    if ! run_unit_tests; then
        failed_tests+=("单元测试")
    fi
    
    # 集成测试
    if ! run_integration_tests; then
        failed_tests+=("集成测试")
    fi
    
    # 基准测试
    if ! run_benchmark_tests; then
        failed_tests+=("基准测试")
    fi
    
    # 压力测试
    if ! run_stress_tests; then
        failed_tests+=("压力测试")
    fi
    
    # 生成报告
    generate_test_report
    
    # 计算运行时间
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local formatted_duration=$(printf "%02d:%02d:%02d" $((duration/3600)) $((duration%3600/60)) $((duration%60)))
    
    # 显示结果摘要
    echo
    log_info "===== 测试结果摘要 ====="
    log_info "运行时间: $formatted_duration"
    log_info "输出目录: $OUTPUT_DIR"
    
    if [[ ${#failed_tests[@]} -eq 0 ]]; then
        log_success "所有测试套件都通过了！"
        exit 0
    else
        log_error "以下测试套件失败: ${failed_tests[*]}"
        exit 1
    fi
}

# 运行主函数
main "$@"