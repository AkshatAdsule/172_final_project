import { useEffect, useState } from 'react';
import type { RideSummary, RideDetail } from '../types';
import { ApiService } from '../services/api';

export function useRides() {
    const [rides, setRides] = useState<RideSummary[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchRides = async (page = 1, limit = 50, date?: string) => {
        try {
            setLoading(true);
            setError(null);
            const ridesData = await ApiService.getRides(page, limit, date);
            setRides(ridesData);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to fetch rides');
            console.error('Error fetching rides:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchRides();
    }, []);

    return {
        rides,
        loading,
        error,
        refetch: fetchRides
    };
}

export function useRideDetail(rideId: number | null) {
    const [rideDetail, setRideDetail] = useState<RideDetail | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (rideId === null) {
            setRideDetail(null);
            return;
        }

        const fetchRideDetail = async () => {
            try {
                setLoading(true);
                setError(null);
                const detail = await ApiService.getRideDetail(rideId);
                setRideDetail(detail);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to fetch ride detail');
                console.error('Error fetching ride detail:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchRideDetail();
    }, [rideId]);

    return {
        rideDetail,
        loading,
        error
    };
} 