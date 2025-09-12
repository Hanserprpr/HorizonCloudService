#!/usr/bin/env node

// 验证所有修复是否正确应用
import fs from 'fs';
import path from 'path';

console.log('🔍 验证所有API测试修复...');
console.log('============================================================');

let allFixesApplied = true;
const issues = [];

// 1. 验证comprehensive-api-test.js的修复
console.log('📋 检查1: comprehensive-api-test.js配置修复');
try {
  const testFile = fs.readFileSync('comprehensive-api-test.js', 'utf8');
  
  // 检查文件服务URL配置
  if (testFile.includes("FILE_SERVICE_URL = 'http://localhost:8002'")) {
    console.log('   ✅ 文件服务URL配置正确 (端口8002)');
  } else {
    console.log('   ❌ 文件服务URL配置错误');
    issues.push('文件服务URL应该设置为http://localhost:8002');
    allFixesApplied = false;
  }
  
  // 检查批量操作错误处理
  if (testFile.includes('if (!firstFile || !firstFile.id)')) {
    console.log('   ✅ 批量操作错误处理已修复');
  } else {
    console.log('   ❌ 批量操作错误处理未修复');
    issues.push('批量操作需要添加文件对象验证');
    allFixesApplied = false;
  }
  
  // 检查文件夹名称唯一性
  if (testFile.includes('完整测试文件夹_${timestamp}')) {
    console.log('   ✅ 文件夹名称唯一性已修复');
  } else {
    console.log('   ❌ 文件夹名称唯一性未修复');
    issues.push('文件夹名称需要添加时间戳确保唯一性');
    allFixesApplied = false;
  }
  
} catch (error) {
  console.log('   ❌ 无法读取comprehensive-api-test.js文件');
  issues.push('comprehensive-api-test.js文件不存在或无法读取');
  allFixesApplied = false;
}

// 2. 验证文件服务配置修复
console.log('\n📋 检查2: 文件服务配置修复');
try {
  const configFile = fs.readFileSync('services/file-service/internal/config/config.go', 'utf8');
  
  if (configFile.includes('getEnv("USER_SERVICE_BASE_URL", "http://localhost:8001")')) {
    console.log('   ✅ 用户服务URL默认值已修复');
  } else {
    console.log('   ❌ 用户服务URL默认值未修复');
    issues.push('config.go中USER_SERVICE_BASE_URL默认值应该是http://localhost:8001');
    allFixesApplied = false;
  }
  
} catch (error) {
  console.log('   ❌ 无法读取文件服务配置文件');
  issues.push('services/file-service/internal/config/config.go文件不存在或无法读取');
  allFixesApplied = false;
}

// 3. 验证环境配置文件
console.log('\n📋 检查3: 环境配置文件');
const envFiles = [
  'services/file-service/.env',
  'services/file-service/.env.development'
];

envFiles.forEach(envFile => {
  try {
    if (fs.existsSync(envFile)) {
      const content = fs.readFileSync(envFile, 'utf8');
      if (content.includes('USER_SERVICE_BASE_URL=http://localhost:8001')) {
        console.log(`   ✅ ${envFile} 配置正确`);
      } else {
        console.log(`   ⚠️  ${envFile} 存在但配置可能不完整`);
      }
    } else {
      console.log(`   ⚠️  ${envFile} 不存在（可选）`);
    }
  } catch (error) {
    console.log(`   ❌ 无法读取 ${envFile}`);
  }
});

// 4. 验证辅助工具文件
console.log('\n📋 检查4: 辅助工具文件');
const toolFiles = [
  'test-config.js',
  'API_TEST_FIXES.md',
  'services/file-service/test_fix_verification.go'
];

toolFiles.forEach(toolFile => {
  if (fs.existsSync(toolFile)) {
    console.log(`   ✅ ${toolFile} 存在`);
  } else {
    console.log(`   ❌ ${toolFile} 不存在`);
    issues.push(`缺少辅助工具文件: ${toolFile}`);
  }
});

// 5. 验证启动脚本
console.log('\n📋 检查5: 启动脚本');
const startScripts = [
  'start-services-fixed.sh',
  'start-services-fixed.bat',
  'services/file-service/start_fixed_service.go'
];

let hasStartScript = false;
startScripts.forEach(script => {
  if (fs.existsSync(script)) {
    console.log(`   ✅ ${script} 存在`);
    hasStartScript = true;
  }
});

if (!hasStartScript) {
  console.log('   ⚠️  没有找到启动脚本，但可以手动启动服务');
}

// 总结
console.log('\n============================================================');
if (allFixesApplied && issues.length === 0) {
  console.log('🎉 所有修复都已正确应用！');
  console.log('\n📋 下一步操作:');
  console.log('1. 启动用户服务: cd services/user-service && go run cmd/main.go');
  console.log('2. 启动文件服务: cd services/file-service && go run start_fixed_service.go');
  console.log('3. 运行测试验证: node test-config.js');
  console.log('4. 运行完整测试: node comprehensive-api-test.js');
  console.log('\n预期结果: 测试成功率应该达到100% (33/33)');
} else {
  console.log('⚠️  发现一些问题需要解决:');
  issues.forEach((issue, index) => {
    console.log(`${index + 1}. ${issue}`);
  });
  console.log('\n请根据API_TEST_FIXES.md文档中的说明进行修复。');
}

console.log('\n📚 详细修复说明请参考: API_TEST_FIXES.md');
console.log('============================================================');
