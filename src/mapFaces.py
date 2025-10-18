import csv
import os
from pathlib import Path
import shutil
import unicodedata
import logging

logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)

DELIMITER = ";"
ENCODING = "utf-8-sig"


class PlayerMapping:
    src_player_id: str
    dest_player_id: str

    def __init__(self, src_player_id: str, dest_player_id: str):
        self.src_player_id = src_player_id
        self.dest_player_id = dest_player_id


def read_csv(file_path: str):
    logger.info(f"Reading CSV file: {file_path}")
    data = {}
    try:
        with open(file_path, mode="r", encoding=ENCODING) as file:
            reader = csv.DictReader(file, delimiter=DELIMITER)
            for row in reader:
                player_id = row["Id"]
                player_name = row["Name"]
                data[player_name] = player_id
        logger.info(f"Successfully read {len(data)} players from {file_path}")
    except Exception as e:
        logger.error(f"Error reading CSV file {file_path}: {e}")
        raise
    return data


def get_player_mapping(source_csv: str, destination_csv: str) -> list[PlayerMapping]:
    logger.info("Starting player mapping process")
    source_data = read_csv(source_csv)
    destination_data = read_csv(destination_csv)
    player_mapping = []

    # Use a set for O(1) lookup and removal
    available_names = set(destination_data.keys())

    for player_name in source_data.keys():
        candidate_name = get_best_match(player_name, available_names)
        if candidate_name:
            player_mapping.append(
                PlayerMapping(
                    source_data[player_name], destination_data[candidate_name]
                )
            )
            logger.debug(f"Matched '{player_name}' -> '{candidate_name}'")
            # Remove matched name from available pool
            available_names.discard(candidate_name)
        else:
            logger.debug(f"No match found for player: {player_name}")

    logger.info(f"Successfully mapped {len(player_mapping)} players")
    return player_mapping


def update_faces_structure(
    src_folder_path: str, dest_folder_path: str, mapping: list[PlayerMapping]
):
    logger.info(
        f"Updating faces structure from {src_folder_path} to {dest_folder_path}"
    )
    facePath = "Asset/model/character/face/real"
    processed = 0

    for item in mapping:
        src_path = f"{src_folder_path}/{facePath}/{item.src_player_id}"
        dest_path = f"{dest_folder_path}/{facePath}/{item.dest_player_id}"

        if not os.path.exists(src_path):
            logger.warning(f"Source path does not exist: {src_path}")
            continue

        if not os.path.exists(dest_path):
            os.makedirs(dest_path)
            logger.debug(f"Created directory: {dest_path}")

        shutil.copytree(src_path, dest_path, dirs_exist_ok=True)
        hex_replace(
            f"{dest_path}/#Win/face.fpk", item.src_player_id, item.dest_player_id
        )
        processed += 1

    logger.info(f"Successfully processed {processed} player faces")


def hex_replace(file_path, old_id, new_id):
    if not Path(file_path).exists():
        logger.debug(f"File does not exist, skipping hex replace: {file_path}")
        return

    try:
        with open(file_path, "rb") as f:
            data = f.read()

        data = data.replace(old_id.encode(), new_id.encode())

        with open(file_path, "wb") as f:
            f.write(data)
        logger.debug(f"Hex replaced in {file_path}: {old_id} -> {new_id}")
    except Exception as e:
        logger.error(f"Error during hex replace in {file_path}: {e}")


def normalize(fullName: str):
    """Remove accents and non-alphanumeric characters (keep spaces)"""
    nfd = unicodedata.normalize("NFD", fullName)
    cleaned = "".join(
        character
        for character in nfd
        if unicodedata.category(character) != "Mn"
        and (character.isalnum() or character.isspace())
    )
    return cleaned.casefold()


def get_best_match(target_name: str, candidate_names: set[str]):
    """Find best matching name from candidates with early filtering. Returns closest match or None."""
    # Early return for exact match
    if target_name in candidate_names:
        return target_name

    # Normalize target once (cache it)
    target_normalized = normalize(target_name)
    target_parts = target_normalized.split()

    # Extract surname for quick filtering
    target_surname = target_parts[-1] if target_parts else ""

    best_match = None
    best_score = -1

    for candidate in candidate_names:
        # Quick checks before expensive normalization

        # 1. Filter by approximate length (allow some flexibility)
        if abs(len(candidate) - len(target_name)) > 10:
            continue

        # 2. Check if surnames share first character (cheap check)
        if target_surname and not any(
            word.lower().startswith(target_surname[0]) for word in candidate.split()
        ):
            continue

        # Now do the expensive matching
        score = calculate_name_match_score(target_parts, target_normalized, candidate)
        if score is not None and score > best_score:
            best_score = score
            best_match = candidate

    return best_match


def calculate_name_match_score(
    target_parts: list[str], target_normalized: str, candidate_name: str
):
    """Optimized matching that accepts pre-normalized target. Returns None if no match."""
    # Normalize candidate
    candidate_normalized = normalize(candidate_name)

    # Quick exact match after normalization
    if target_normalized == candidate_normalized:
        return 1000  # Perfect match gets highest score

    candidate_parts = candidate_normalized.split()

    # Length mismatch - early exit
    if len(target_parts) != len(candidate_parts):
        return None

    if len(target_parts) == 1:
        return 100 if target_parts[0] == candidate_parts[0] else None

    # Surname must match exactly - check first to fail fast
    if target_parts[-1] != candidate_parts[-1]:
        return None

    score = 0

    # Check all first names
    for p1, p2 in zip(target_parts[:-1], candidate_parts[:-1]):
        if len(p1) == 1 or len(p2) == 1:
            if p1[0] != p2[0]:
                return None
            score += 1
        else:
            if p1 != p2:
                return None
            score += 10

    score += 50
    return score


if __name__ == "__main__":
    logger.info("=== Starting Player Face Mapping Tool ===")

    source_csv = "samples/BPB-2023-players.csv"
    destination_csv = "samples/FL26_players.csv"
    src_folder_path = "samples"
    dest_folder_path = "livecpk/root"

    try:
        mapping = get_player_mapping(source_csv, destination_csv)
        update_faces_structure(src_folder_path, f"result/{dest_folder_path}", mapping)
        logger.info("=== Finished processing successfully ===")
    except Exception as e:
        logger.critical(f"Fatal error during processing: {e}", exc_info=True)
        raise
