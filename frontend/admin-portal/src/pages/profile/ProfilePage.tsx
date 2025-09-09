import React from 'react';
import { Typography, Card } from 'antd';

const { Title } = Typography;

const ProfilePage: React.FC = () => {
  return (
    <div>
      <Title level={2}>个人资料</Title>
      
      <Card>
        <p>用户个人资料编辑将在这里显示</p>
        <p>包括基本信息、密码修改、偏好设置等</p>
      </Card>
    </div>
  );
};

export default ProfilePage;