export interface LatLng {
	lat: number;
	lng: number;
	speed_knots?: number;
}

export interface Position {
	latitude: number;
	longitude: number;
	speed_knots?: number;
	timestamp: string; // ISO 8601 format
}

export interface RideSummary {
	id: number;
	name: string;
	start_time: string; // ISO 8601 format
	end_time?: string; // ISO 8601 format, optional for ongoing rides
}

export interface RideDetail {
	id: number;
	name: string;
	start_time: string;
	end_time?: string;
	positions: Position[];
}

// WebSocket message types based on planning document
export interface WebSocketMessage {
	type:
		| "current_location"
		| "RIDE_STARTED"
		| "RIDE_POSITION_UPDATE"
		| "RIDE_ENDED";
	payload: any;
	timestamp: string;
}

export interface CurrentLocationMessage extends WebSocketMessage {
	type: "current_location";
	payload: {
		latitude: number;
		longitude: number;
		timestamp: string;
		speed_knots?: number;
	};
}

export interface RideStartedMessage extends WebSocketMessage {
	type: "RIDE_STARTED";
	payload: {
		ride_id: number;
		ride_name: string;
		timestamp: string;
		position: Position;
	};
}

export interface RidePositionUpdateMessage extends WebSocketMessage {
	type: "RIDE_POSITION_UPDATE";
	payload: {
		ride_id: number;
		timestamp: string;
		position: Position;
	};
}

export interface RideEndedMessage extends WebSocketMessage {
	type: "RIDE_ENDED";
	payload: {
		ride_id: number;
		timestamp: string;
	};
}
