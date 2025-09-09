import { RouterProvider } from 'react-router-dom';
import { stableRouter } from './router/stable';
import ErrorBoundary from '@components/common/ErrorBoundary';
import GlobalErrorHandler from '@components/common/GlobalErrorHandler';

function App() {
  return (
    <ErrorBoundary>
      <GlobalErrorHandler>
        <RouterProvider router={stableRouter} />
      </GlobalErrorHandler>
    </ErrorBoundary>
  );
}

export default App;
