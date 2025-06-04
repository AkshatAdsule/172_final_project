// Convert knots to miles per hour
export function knotsToMph(knots: number): number {
	return knots * 1.15078;
}

// Format speed in mph with one decimal place
export function formatSpeedMph(knots: number): string {
	return knotsToMph(knots).toFixed(1);
}
