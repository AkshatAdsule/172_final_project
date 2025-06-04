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
		| "ride_started"
		| "ride_position_added"
		| "ride_ended";
	payload: any;
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
	type: "ride_started";
	payload: {
		ride_id: number;
		name: string;
		start_time: string;
		initial_position: Position;
	};
}

export interface RidePositionAddedMessage extends WebSocketMessage {
	type: "ride_position_added";
	payload: {
		ride_id: number;
		position: Position;
	};
}

export interface RideEndedMessage extends WebSocketMessage {
	type: "ride_ended";
	payload: {
		ride_id: number;
		end_time: string;
	};
}
