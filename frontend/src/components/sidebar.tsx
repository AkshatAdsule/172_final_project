import { Link } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { useLockStatus } from "../hooks/useLockStatus";
import { useRides } from "../hooks/useRides";
import { getEffectiveEndTime } from "../utils/rideUtils";
import { Button } from "./button";
import styles from "./sidebar.module.css";

export default function Sidebar() {
	const { rides, loading, error } = useRides();
	const { lockStatus, isToggling, toggleLockStatus } = useLockStatus();
	const ridesRef = useRef<HTMLDivElement>(null);
	const [isScrolling, setIsScrolling] = useState(false);
	const scrollTimeoutRef = useRef<number | undefined>(undefined);

	// Handle scroll events for auto-hide scrollbar
	useEffect(() => {
		const ridesElement = ridesRef.current;
		if (!ridesElement) return;

		const handleScroll = () => {
			setIsScrolling(true);

			// Clear existing timeout
			if (scrollTimeoutRef.current !== undefined) {
				window.clearTimeout(scrollTimeoutRef.current);
			}

			// Set timeout to hide scrollbar after 1.5 seconds of no scrolling
			scrollTimeoutRef.current = window.setTimeout(() => {
				setIsScrolling(false);
			}, 1500);
		};

		ridesElement.addEventListener("scroll", handleScroll);

		return () => {
			ridesElement.removeEventListener("scroll", handleScroll);
			if (scrollTimeoutRef.current !== undefined) {
				window.clearTimeout(scrollTimeoutRef.current);
			}
		};
	}, []);

	const formatDate = (dateString: string) => {
		return new Date(dateString).toLocaleDateString();
	};

	const formatTime = (dateString: string) => {
		return new Date(dateString).toLocaleTimeString([], {
			hour: "2-digit",
			minute: "2-digit",
		});
	};

	return (
		<nav className={styles.nav}>
			<div className={styles.brand}>
				<h1>
					B<sup>3</sup>
				</h1>
			</div>
			<hr />
			<div className={styles.lockSection}>
				<Button
					variant={lockStatus === "LOCKED" ? "destructive" : "default"}
					onClick={toggleLockStatus}
					disabled={isToggling}
					className={styles.lockButton}
				>
					{isToggling
						? "..."
						: lockStatus === "LOCKED"
							? "🔒 Locked"
							: "🔓 Unlocked"}
				</Button>
			</div>
			<hr />
			<div className={styles.ridesHeader}>
				<Link to="/rides/live">
					<div>Live Track</div>
				</Link>
				<hr />
			</div>
			<div
				ref={ridesRef}
				className={`${styles.rides} ${isScrolling ? styles.scrolling : ""}`}
			>
				{loading && <div className={styles.loading}>Loading rides...</div>}
				{error && <div className={styles.error}>Error: {error}</div>}

				{!loading && !error && rides.length === 0 && (
					<div className={styles.empty}>No rides found</div>
				)}

				{!loading &&
					!error &&
					rides.map((ride) => (
						<Link
							key={ride.id}
							to={"/rides/$rideId"}
							params={{ rideId: ride.id.toString() }}
							activeProps={{
								className: styles.activeRide,
							}}
						>
							<div className={styles.ride}>
								<span className={styles.name}>{ride.name}</span>
								<span className={styles.date}>
									{formatDate(ride.start_time)}
								</span>
								<span className={styles.time}>
									{formatTime(ride.start_time)}
									{(() => {
										const effectiveEndTime = getEffectiveEndTime(ride);
										return effectiveEndTime
											? ` - ${formatTime(effectiveEndTime)}`
											: "";
									})()}
								</span>
							</div>
						</Link>
					))}
			</div>
		</nav>
	);
}
