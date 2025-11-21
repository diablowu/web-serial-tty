import { useEffect, useRef, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';

export default function TerminalView() {
    const { deviceId } = useParams();
    const navigate = useNavigate();
    const terminalRef = useRef(null);
    const wsRef = useRef(null);
    const xtermRef = useRef(null);
    const fitAddonRef = useRef(null);

    const [inputMsg, setInputMsg] = useState('');
    const [inputMode, setInputMode] = useState('ASCII'); // 'ASCII' | 'HEX'
    const [autoScroll, setAutoScroll] = useState(true);

    // Helper to format time
    const getTimeStr = () => {
        const now = new Date();
        return `[${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}.${now.getMilliseconds().toString().padStart(3, '0')}] `;
    };

    // Helper to write to terminal with color and timestamp
    const writeToTerminal = useCallback((msg, type) => {
        if (!xtermRef.current) return;

        const term = xtermRef.current;
        const timeStr = getTimeStr();

        // Color codes: 
        // RX (Received): Cyan \x1b[36m
        // TX (Sent): Green \x1b[32m
        // Reset: \x1b[0m
        const color = type === 'TX' ? '\x1b[32m' : '\x1b[36m';
        const prefix = `\r\n${color}${timeStr}${type === 'TX' ? 'TX > ' : 'RX < '}\x1b[0m`;

        // Handle newlines for xterm
        const formattedMsg = msg.replace(/\n/g, '\r\n');

        term.write(prefix + formattedMsg);

        if (autoScroll) {
            term.scrollToBottom();
        }
    }, [autoScroll]);

    useEffect(() => {
        const term = new Terminal({
            cursorBlink: true,
            fontSize: 14,
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            theme: {
                background: '#000000',
                foreground: '#ffffff',
            },
            disableStdin: true, // Make terminal read-only for user input
            convertEol: true,
        });
        const fitAddon = new FitAddon();
        term.loadAddon(fitAddon);

        term.open(terminalRef.current);
        fitAddon.fit();

        xtermRef.current = term;
        fitAddonRef.current = fitAddon;

        // Handle resize
        const handleResize = () => fitAddon.fit();
        window.addEventListener('resize', handleResize);

        // Connect WebSocket
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/client?device_id=${deviceId}`;

        console.log('Connecting to', wsUrl);
        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
            term.write(`\r\n\x1b[1;33mSystem: Connected to ${deviceId}\x1b[0m\r\n`);
        };

        ws.onmessage = (event) => {
            if (typeof event.data === 'string') {
                writeToTerminal(event.data, 'RX');
            } else {
                const reader = new FileReader();
                reader.onload = () => {
                    writeToTerminal(reader.result, 'RX');
                };
                reader.readAsText(event.data);
            }
        };

        ws.onclose = () => {
            term.write('\r\n\x1b[1;31mSystem: Connection closed\x1b[0m\r\n');
        };

        ws.onerror = (err) => {
            console.error('WebSocket error:', err);
            term.write('\r\n\x1b[1;31mSystem: WebSocket error\x1b[0m\r\n');
        };

        return () => {
            window.removeEventListener('resize', handleResize);
            if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
                ws.close();
            }
            term.dispose();
        };
    }, [deviceId, writeToTerminal]);

    const handleSend = () => {
        if (!inputMsg) return;
        if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
            alert('Not connected');
            return;
        }

        if (inputMode === 'HEX') {
            // Convert Hex string to byte array
            const cleanHex = inputMsg.replace(/\s+/g, '');
            if (!/^[0-9A-Fa-f]+$/.test(cleanHex)) {
                alert('Invalid Hex format');
                return;
            }
            if (cleanHex.length % 2 !== 0) {
                alert('Hex string must have an even number of characters');
                return;
            }

            const buffer = new Uint8Array(cleanHex.length / 2);
            for (let i = 0; i < cleanHex.length; i += 2) {
                buffer[i / 2] = parseInt(cleanHex.substring(i, i + 2), 16);
            }

            wsRef.current.send(buffer);
            writeToTerminal(`[HEX] ${inputMsg}`, 'TX');
        } else {
            // ASCII
            wsRef.current.send(inputMsg);
            writeToTerminal(inputMsg, 'TX');
        }

        setInputMsg('');
    };

    const handleClear = () => {
        xtermRef.current?.clear();
    };

    return (
        <div className="terminal-page">
            <div className="terminal-header">
                <div className="header-left">
                    <button className="btn-back" onClick={() => navigate('/')}>Back</button>
                    <span className="device-title">{deviceId}</span>
                </div>
                <div className="header-controls">
                    <label className="checkbox-label">
                        <input
                            type="checkbox"
                            checked={autoScroll}
                            onChange={e => setAutoScroll(e.target.checked)}
                        />
                        <span>Auto Scroll</span>
                    </label>
                    <button className="btn-clear" onClick={handleClear}>Clear</button>
                </div>
            </div>

            <div className="terminal-container" ref={terminalRef} />

            <div className="control-panel">
                <div className="input-row">
                    <select
                        value={inputMode}
                        onChange={e => setInputMode(e.target.value)}
                        className="mode-select"
                    >
                        <option value="ASCII">ASCII</option>
                        <option value="HEX">HEX</option>
                    </select>
                    <input
                        type="text"
                        value={inputMsg}
                        onChange={e => setInputMsg(e.target.value)}
                        onKeyDown={e => e.key === 'Enter' && handleSend()}
                        placeholder={inputMode === 'HEX' ? "e.g. AA BB CC" : "Type message..."}
                        className="msg-input"
                    />
                    <button onClick={handleSend} className="btn-send">Send</button>
                </div>
            </div>
        </div>
    );
}
