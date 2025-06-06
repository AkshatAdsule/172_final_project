import { useEffect, useState } from "react";
import { ApiService } from "../services/api";

export type LockStatus = "LOCKED" | "UNLOCKED";

export function useLockStatus() {
	const [lockStatus, setLockStatus] = useState<LockStatus | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [isToggling, setIsToggling] = useState(false);

	// Fetch initial lock status
	useEffect(() => {
		const fetchLockStatus = async () => {
			try {
				setLoading(true);
				setError(null);
				const response = await ApiService.getLockStatus();
				setLockStatus(response.status);
			} catch (err) {
				setError(
					err instanceof Error ? err.message : "Failed to fetch lock status",
				);
				console.error("Error fetching lock status:", err);
			} finally {
				setLoading(false);
			}
		};

		fetchLockStatus();
	}, []);

	// Toggle lock status
	const toggleLockStatus = async () => {
		if (!lockStatus || isToggling) return;

		const newStatus: LockStatus =
			lockStatus === "LOCKED" ? "UNLOCKED" : "LOCKED";

		try {
			setIsToggling(true);
			setError(null);
			const response = await ApiService.setLockStatus(newStatus);
			setLockStatus(response.status);
		} catch (err) {
			setError(
				err instanceof Error ? err.message : "Failed to update lock status",
			);
			console.error("Error updating lock status:", err);
		} finally {
			setIsToggling(false);
		}
	};

	return {
		lockStatus,
		loading,
		error,
		isToggling,
		toggleLockStatus,
	};
}
