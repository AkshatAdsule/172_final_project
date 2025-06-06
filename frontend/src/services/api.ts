import type { RideDetail, RideSummary } from "../types";

const API_BASE_URL = import.meta.env.VITE_API_BASE
	? `http${import.meta.env.VITE_API_BASE}/api`
	: "http://localhost:8080/api";

export class ApiService {
	private static async request<T>(
		endpoint: string,
		options?: RequestInit,
	): Promise<T> {
		const url = `${API_BASE_URL}${endpoint}`;

		try {
			const response = await fetch(url, {
				headers: {
					"Content-Type": "application/json",
					...options?.headers,
				},
				...options,
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			return await response.json();
		} catch (error) {
			console.error(`API request failed for ${endpoint}:`, error);
			throw error;
		}
	}

	/**
	 * Get all rides summary
	 * @param page - Page number (1-based)
	 * @param limit - Number of rides per page
	 * @param date - Filter by start date (YYYY-MM-DD format)
	 */
	static async getRides(
		page = 1,
		limit = 10,
		date?: string,
	): Promise<RideSummary[]> {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString(),
		});

		if (date) {
			params.append("date", date);
		}

		return this.request<RideSummary[]>(`/rides?${params.toString()}`);
	}

	/**
	 * Get detailed information for a specific ride
	 * @param rideId - The ID of the ride
	 */
	static async getRideDetail(rideId: number): Promise<RideDetail> {
		return this.request<RideDetail>(`/rides/${rideId}`);
	}

	/**
	 * Get the current lock status
	 */
	static async getLockStatus(): Promise<{ status: "LOCKED" | "UNLOCKED" }> {
		return this.request<{ status: "LOCKED" | "UNLOCKED" }>("/getLockStatus");
	}

	/**
	 * Set the lock status
	 * @param status - The desired lock status
	 */
	static async setLockStatus(
		status: "LOCKED" | "UNLOCKED",
	): Promise<{ status: "LOCKED" | "UNLOCKED" }> {
		return this.request<{ status: "LOCKED" | "UNLOCKED" }>("/setLockStatus", {
			method: "POST",
			body: JSON.stringify({ status }),
		});
	}
}
