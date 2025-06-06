import type { RideDetail, RideSummary } from "../types";

/**
 * Get the effective end time for a ride.
 * For ongoing rides or invalid end times, use the latest position timestamp.
 */
export function getEffectiveEndTime(
	ride: RideSummary | RideDetail,
): string | undefined {
	// If there's a valid end time that's after start time, use it
	if (ride.end_time) {
		const startTime = new Date(ride.start_time);
		const endTime = new Date(ride.end_time);
		if (endTime >= startTime) {
			return ride.end_time;
		}
	}

	// For RideDetail, we can use the latest position timestamp
	if ("positions" in ride && ride.positions.length > 0) {
		const latestPosition = ride.positions[ride.positions.length - 1];
		return latestPosition.timestamp;
	}

	// For ongoing rides without position data, return undefined
	return undefined;
}
