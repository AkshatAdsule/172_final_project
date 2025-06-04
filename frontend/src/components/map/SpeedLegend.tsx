import styles from "./SpeedLegend.module.css";
import { formatSpeedMph } from "../../utils/speed";

interface SpeedLegendProps {
	minSpeed: number;
	maxSpeed: number;
}

// Color interpolation function (same as in SpeedPolyline)
function interpolateColor(
	speed: number,
	minSpeed: number,
	maxSpeed: number,
): string {
	const normalizedSpeed = Math.max(
		0,
		Math.min(1, (speed - minSpeed) / (maxSpeed - minSpeed)),
	);

	const r = Math.round(normalizedSpeed * 255 + (1 - normalizedSpeed) * 0);
	const g = Math.round(normalizedSpeed * 82 + (1 - normalizedSpeed) * 123);
	const b = Math.round(normalizedSpeed * 82 + (1 - normalizedSpeed) * 255);

	return `rgb(${r}, ${g}, ${b})`;
}

export function SpeedLegend({ minSpeed, maxSpeed }: SpeedLegendProps) {
	if (minSpeed === maxSpeed) return null;

	const steps = 5;
	const labels = [];

	// Create gradient stops for the CSS gradient (from bottom to top)
	const gradientStops = [];
	for (let i = 0; i <= 20; i++) {
		const speed = minSpeed + (maxSpeed - minSpeed) * (i / 20);
		const color = interpolateColor(speed, minSpeed, maxSpeed);
		const percentage = (i / 20) * 100;
		gradientStops.push(`${color} ${percentage}%`);
	}

	// Create labels (from max at top to min at bottom)
	for (let i = 0; i <= steps; i++) {
		const speed = maxSpeed - (maxSpeed - minSpeed) * (i / steps); // Reversed: start from max
		const position = (i / steps) * 100;
		labels.push({
			speed: formatSpeedMph(speed),
			position,
		});
	}

	const gradientStyle = {
		background: `linear-gradient(to top, ${gradientStops.join(", ")})`,
	};

	return (
		<div className={styles.speedLegend}>
			<div className={styles.title}>Speed (mph)</div>
			<div className={styles.gradientContainer}>
				<div className={styles.gradientBar} style={gradientStyle} />
				<div className={styles.labels}>
					{labels.map((label, index) => (
						<div
							key={index}
							className={styles.label}
							style={{ top: `${label.position}%` }}
						>
							{label.speed}
						</div>
					))}
				</div>
			</div>
		</div>
	);
}
