import { RouterProvider } from 'react-router-dom';
import { App as AntdApp } from 'antd';
import { stableRouter } from './router/stable';
import ErrorBoundary from '@components/common/ErrorBoundary';
import GlobalErrorHandler from '@components/common/GlobalErrorHandler';

function App() {
  return (
    <ErrorBoundary>
      <GlobalErrorHandler>
        <AntdApp>
          <RouterProvider router={stableRouter} />
        </AntdApp>
      </GlobalErrorHandler>
    </ErrorBoundary>
  );
}

export default App;
