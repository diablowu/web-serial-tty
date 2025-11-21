import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import DeviceList from './components/DeviceList';
import TerminalView from './components/TerminalView';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<DeviceList />} />
        <Route path="/terminal/:deviceId" element={<TerminalView />} />
      </Routes>
    </Router>
  );
}

export default App;
