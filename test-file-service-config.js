#!/usr/bin/env node

// 测试当前运行的文件服务配置
import axios from 'axios';

const FILE_SERVICE_URL = 'http://localhost:8002';

async function testFileServiceConfig() {
  console.log('🔍 测试当前运行的文件服务配置...');
  
  try {
    // 1. 测试健康检查
    console.log('📋 测试文件服务健康检查...');
    const healthResponse = await axios.get(`${FILE_SERVICE_URL}/api/v1/health`);
    console.log('✅ 文件服务健康检查成功:', healthResponse.status);
    
    // 2. 创建测试用户并获取token
    console.log('📋 创建测试用户...');
    const timestamp = Date.now().toString().slice(-6);
    const testUser = {
      student_id: `test_config_${timestamp}`,
      password: 'TestConfig123!',
      nick_name: '配置测试用户',
      email: `test_config_${timestamp}@example.com`
    };
    
    const registerResponse = await axios.post('http://localhost:8001/api/v1/auth/register', testUser);
    console.log('✅ 用户注册成功:', registerResponse.status);
    
    const loginResponse = await axios.post('http://localhost:8001/api/v1/auth/login', {
      student_id: testUser.student_id,
      password: testUser.password
    });
    console.log('✅ 用户登录成功:', loginResponse.status);
    
    const accessToken = loginResponse.data.data.access_token;
    const userId = loginResponse.data.data.user.id;
    console.log('🔑 获取到访问令牌和用户ID:', userId);
    
    // 3. 测试配额检查 - 这是失败的关键测试
    console.log('📋 测试配额检查 (关键测试)...');
    try {
      const quotaResponse = await axios.get(`${FILE_SERVICE_URL}/api/v1/quota/status`, {
        headers: { 'Authorization': `Bearer ${accessToken}` },
        timeout: 5000  // 5秒超时，避免长时间等待
      });
      console.log('✅ 配额检查成功:', quotaResponse.status);
      console.log('📊 配额信息:', JSON.stringify(quotaResponse.data, null, 2));
    } catch (error) {
      console.log('❌ 配额检查失败:', error.message);
      if (error.response) {
        console.log('   状态码:', error.response.status);
        console.log('   错误详情:', JSON.stringify(error.response.data, null, 2));
        
        // 分析错误信息
        const errorMessage = error.response.data?.message || '';
        if (errorMessage.includes('unsupported protocol scheme')) {
          console.log('🔍 分析: 这是用户服务URL配置问题');
          console.log('   - 文件服务无法正确调用用户服务API');
          console.log('   - 可能的原因: USER_SERVICE_BASE_URL环境变量为空或格式错误');
        }
      }
    }
    
    // 4. 测试简单文件上传 - 另一个失败的测试
    console.log('📋 测试简单文件上传...');
    try {
      const formData = new FormData();
      const testContent = 'This is a test file content for configuration testing.';
      const blob = new Blob([testContent], { type: 'text/plain' });
      formData.append('file', blob, 'config-test.txt');
      formData.append('folder_id', '');
      
      const uploadResponse = await axios.post(`${FILE_SERVICE_URL}/api/v1/files/upload`, formData, {
        headers: { 
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'multipart/form-data'
        },
        timeout: 5000
      });
      console.log('✅ 文件上传成功:', uploadResponse.status);
    } catch (error) {
      console.log('❌ 文件上传失败:', error.message);
      if (error.response) {
        console.log('   状态码:', error.response.status);
        console.log('   错误详情:', JSON.stringify(error.response.data, null, 2));
      }
    }
    
  } catch (error) {
    console.log('💥 测试过程中发生错误:', error.message);
  }
}

testFileServiceConfig().then(() => {
  console.log('🎉 配置测试完成');
}).catch(error => {
  console.error('💥 测试失败:', error.message);
});
