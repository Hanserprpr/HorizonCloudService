#!/usr/bin/env node

// 简单的配置测试脚本
import axios from 'axios';

const USER_SERVICE_URL = 'http://localhost:8001';
const FILE_SERVICE_URL = 'http://localhost:8002';

async function testServices() {
  console.log('🔍 测试服务连接...');
  
  try {
    // 测试用户服务
    console.log('📋 测试用户服务连接...');
    const userHealthResponse = await axios.get(`${USER_SERVICE_URL}/health`, { timeout: 5000 });
    console.log('✅ 用户服务连接正常:', userHealthResponse.status);
  } catch (error) {
    console.log('❌ 用户服务连接失败:', error.message);
  }
  
  try {
    // 测试文件服务
    console.log('📋 测试文件服务连接...');
    const fileHealthResponse = await axios.get(`${FILE_SERVICE_URL}/api/v1/health`, { timeout: 5000 });
    console.log('✅ 文件服务连接正常:', fileHealthResponse.status);
  } catch (error) {
    console.log('❌ 文件服务连接失败:', error.message);
  }
  
  // 测试用户注册和登录
  try {
    console.log('📋 测试用户注册...');
    const timestamp = Date.now().toString().slice(-6);
    const testUser = {
      student_id: `test_config_${timestamp}`,
      password: 'TestConfig123!',
      nick_name: '配置测试用户',
      email: `test_config_${timestamp}@example.com`
    };
    
    const registerResponse = await axios.post(`${USER_SERVICE_URL}/api/v1/auth/register`, testUser);
    console.log('✅ 用户注册成功:', registerResponse.status);
    
    console.log('📋 测试用户登录...');
    const loginResponse = await axios.post(`${USER_SERVICE_URL}/api/v1/auth/login`, {
      student_id: testUser.student_id,
      password: testUser.password
    });
    console.log('✅ 用户登录成功:', loginResponse.status);
    
    const accessToken = loginResponse.data.data.access_token;
    console.log('🔑 获取到访问令牌:', accessToken.substring(0, 20) + '...');
    
    // 测试文件服务认证
    console.log('📋 测试文件服务认证...');
    const fileAuthResponse = await axios.get(`${FILE_SERVICE_URL}/api/v1/auth/me`, {
      headers: { 'Authorization': `Bearer ${accessToken}` }
    });
    console.log('✅ 文件服务认证成功:', fileAuthResponse.status);
    
    // 测试配额检查
    console.log('📋 测试配额检查...');
    const quotaResponse = await axios.get(`${FILE_SERVICE_URL}/api/v1/quota/status`, {
      headers: { 'Authorization': `Bearer ${accessToken}` }
    });
    console.log('✅ 配额检查成功:', quotaResponse.status);
    
  } catch (error) {
    console.log('❌ 认证测试失败:', error.message);
    if (error.response) {
      console.log('   响应状态:', error.response.status);
      console.log('   响应数据:', JSON.stringify(error.response.data, null, 2));
    }
  }
}

testServices().then(() => {
  console.log('🎉 配置测试完成');
}).catch(error => {
  console.error('💥 测试过程中发生错误:', error.message);
});
