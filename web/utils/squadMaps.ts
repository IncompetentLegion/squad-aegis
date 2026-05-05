const SQUAD_MAPS_THUMBNAIL_BASE_URL =
    "https://raw.githubusercontent.com/mahtoid/SquadMaps/refs/heads/master/img/maps/thumbnails";

function addCandidate(candidates: string[], value: string) {
    if (value && !candidates.includes(value)) {
        candidates.push(value);
    }
}

function stripModLayerAffixes(layer: string): string {
    return layer
        .replace(/^(?:SU_|SPM_)+/, "")
        .replace(/_(?:PreCap|HalfCap)$/i, "");
}

function normalizeLayerVersion(layer: string): string {
    return layer.replace(/_V(\d+)$/i, "_v$1");
}

function zeroPadLayerVersion(layer: string): string {
    return layer.replace(/_v([1-9])$/i, "_v0$1");
}

function stripSeedVehicle(layer: string): string {
    return layer.replace(/_Seed_Vehicle_v/i, "_Seed_v");
}

export function getSquadMapsThumbnailCandidates(layer?: string): string[] {
    if (!layer) {
        return [];
    }

    const candidates: string[] = [];
    const trimmed = layer.trim();
    const normalized = normalizeLayerVersion(stripModLayerAffixes(trimmed));
    const withoutSeedVehicle = stripSeedVehicle(normalized);

    addCandidate(candidates, trimmed);
    addCandidate(candidates, normalized);
    addCandidate(candidates, zeroPadLayerVersion(normalized));
    addCandidate(candidates, withoutSeedVehicle);
    addCandidate(candidates, zeroPadLayerVersion(withoutSeedVehicle));

    return candidates;
}

export function getSquadMapsThumbnailUrlForCandidate(candidate: string): string {
    return `${SQUAD_MAPS_THUMBNAIL_BASE_URL}/${encodeURIComponent(candidate)}.jpg`;
}
