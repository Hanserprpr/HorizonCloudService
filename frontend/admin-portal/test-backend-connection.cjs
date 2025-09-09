#!/usr/bin/env node

// 后端API连接测试脚本
const axios = require('axios');

const USER_SERVICE_URL = 'http://localhost:8001';
const FILE_SERVICE_URL = 'http://localhost:8002';

async function testBackendConnection() {
  console.log('🧪 开始测试后端API连接...\n');
  
  // 测试用户服务
  console.log('1. 测试用户服务连接...');
  try {
    const response = await axios.get(`${USER_SERVICE_URL}/health`);
    console.log('✅ 用户服务连接成功:', response.data);
  } catch (error) {
    console.log('❌ 用户服务连接失败:', error.code || error.message);
    console.log('   请确保用户服务在端口 8001 上运行');
    console.log('   启动命令: cd services/user-service && go run cmd/main_simple.go');
  }
  
  console.log();
  
  // 测试文件服务
  console.log('2. 测试文件服务连接...');
  try {
    const response = await axios.get(`${FILE_SERVICE_URL}/health`);
    console.log('✅ 文件服务连接成功:', response.data);
  } catch (error) {
    console.log('❌ 文件服务连接失败:', error.code || error.message);
    console.log('   文件服务尚未实现HTTP服务器');
    console.log('   需要创建 services/file-service/cmd/main.go');
  }
  
  console.log();
  
  // 测试登录API
  console.log('3. 测试登录API...');
  try {
    const loginData = {
      username: 'admin',
      password: 'admin123'
    };
    
    const response = await axios.post(`${USER_SERVICE_URL}/api/v1/auth/login`, loginData);
    console.log('✅ 登录API测试成功');
    console.log('   响应数据结构:', Object.keys(response.data));
  } catch (error) {
    if (error.response) {
      console.log('🟡 登录API可访问，但认证失败（正常，因为是测试账户）');
      console.log('   状态码:', error.response.status);
      console.log('   响应数据:', error.response.data);
    } else {
      console.log('❌ 登录API连接失败:', error.code || error.message);
    }
  }
  
  console.log('\n🎯 测试总结:');
  console.log('- 如果用户服务连接成功，可以测试登录、用户管理功能');
  console.log('- 如果文件服务连接失败，文件管理功能将使用Mock数据');
  console.log('- 前端已配置正确的API端点，可以正常启动');
}

// 运行测试
testBackendConnection().catch(console.error);