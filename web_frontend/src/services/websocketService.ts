// WebSocket service for connecting to the PC-side analysis engine

const WEBSOCKET_URL = 'ws://localhost:8080/ws'; // 假设PC端引擎的WebSocket服务地址和端口

let socket: WebSocket | null = null;

interface MessageListener {
  (data: any): void;
}

const listeners: MessageListener[] = [];

export const connectWebSocket = (
  onOpen: () => void,
  onClose: () => void,
  onError: (event: Event) => void
) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    console.log('WebSocket already connected.');
    onOpen();
    return;
  }

  socket = new WebSocket(WEBSOCKET_URL);

  socket.onopen = () => {
    console.log('WebSocket connection established.');
    onOpen();
  };

  socket.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data as string);
      listeners.forEach(listener => listener(data));
    } catch (error) {
      console.error('Error parsing WebSocket message:', error);
    }
  };

  socket.onclose = () => {
    console.log('WebSocket connection closed.');
    socket = null;
    onClose();
  };

  socket.onerror = (event) => {
    console.error('WebSocket error:', event);
    onError(event);
  };
};

export const sendMessage = (message: any) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(message));
  } else {
    console.error('WebSocket is not connected.');
  }
};

export const addMessageListener = (listener: MessageListener) => {
  listeners.push(listener);
};

export const removeMessageListener = (listener: MessageListener) => {
  const index = listeners.indexOf(listener);
  if (index > -1) {
    listeners.splice(index, 1);
  }
};

export const getWebSocketState = () => {
  return socket ? socket.readyState : WebSocket.CLOSED;
};

// Example control command structure (to be refined based on PC engine specs)
export interface ControlCommand {
  action: 'start_capture' | 'stop_capture' | 'set_channel' | 'set_bandwidth';
  payload?: any; // e.g., { channel: number } or { bandwidth: string }
}