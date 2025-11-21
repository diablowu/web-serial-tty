import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

export default function DeviceList() {
    const [devices, setDevices] = useState([]);
    const navigate = useNavigate();

    const fetchDevices = () => {
        fetch('/api/devices')
            .then(res => res.json())
            .then(data => setDevices(data || []))
            .catch(err => console.error('Failed to fetch devices:', err));
    };

    useEffect(() => {
        fetchDevices();
        const interval = setInterval(fetchDevices, 2000); // Poll every 2s
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="container">
            <h1>Web Serial TTY</h1>
            <div className="device-list">
                {devices.length === 0 ? (
                    <p style={{ textAlign: 'center' }}>No devices connected. Waiting...</p>
                ) : (
                    devices.map(id => (
                        <div key={id} className="device-item">
                            <span>{id}</span>
                            <button onClick={() => navigate(`/terminal/${id}`)}>Connect</button>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}
