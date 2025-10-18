import csv
from dataclasses import dataclass
import os
from pathlib import Path
import shutil
import unicodedata
import logging
import argparse


logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)


DELIMITER = ";"
ENCODING = "utf-8-sig"
FACE_PATH = "Asset/model/character/face/real"


@dataclass(frozen=True)
class Player:
    id: str
    name: str


@dataclass(frozen=True)
class PlayerMapping:
    src_player_id: str
    dest_player_id: str


def parse_arguments():
    parser = argparse.ArgumentParser(
        description="Player Face Mapping Tool - Map player faces between game versions"
    )

    parser.add_argument(
        "--source-csv",
        help="Source CSV file with player names and IDs (optional - uses folder names if not provided)",
    )

    parser.add_argument(
        "--destination-csv",
        required=True,
        help="Destination CSV file with player names and IDs (always required)",
    )

    parser.add_argument(
        "--source-folder",
        required=True,
        help="Source folder containing player face directories",
    )

    parser.add_argument(
        "--dest-folder",
        required=True,
        help="Destination folder for mapped player faces",
    )

    return parser.parse_args()


def read_csv(file_path: str) -> list[Player]:
    logger.info(f"Reading CSV file: {file_path}")
    data: list[Player] = []
    try:
        with open(file_path, mode="r", encoding=ENCODING) as file:
            reader = csv.DictReader(file, delimiter=DELIMITER)
            for row in reader:
                player_id = row["Id"]
                player_name = row["Name"]
                data.append(Player(player_id, player_name))
        logger.info(f"Successfully read {len(data)} players from {file_path}")
    except Exception as e:
        logger.error(f"Error reading CSV file {file_path}: {e}")
        raise
    return data


def read_folders(folder_path: str) -> list[Player]:
    """Read folder names as player data, using folder name as both ID and name."""
    logger.info(f"Reading player folders from: {folder_path}")
    data: list[Player] = []

    try:
        folder_path_obj = Path(folder_path)
        if not folder_path_obj.exists():
            logger.error(f"Folder does not exist: {folder_path}")
            raise FileNotFoundError(f"Folder does not exist: {folder_path}")

        for folder in sorted(folder_path_obj.iterdir()):
            if folder.is_dir():
                # Use folder name as both player name and ID
                player_name = folder.name
                player_id = folder.name
                data.append(Player(player_id, player_name))

        logger.info(f"Successfully read {len(data)} player folders")
    except Exception as e:
        logger.error(f"Error reading folders from {folder_path}: {e}")
        raise

    return data


def get_source_data(source_csv: str, source_folder: str) -> list[Player]:
    """Get source player data from CSV or folder names."""
    if source_csv and Path(source_csv).exists():
        logger.info(f"Using CSV mode - reading from {source_csv}")
        return read_csv(source_csv)
    else:
        logger.info(f"Using folder mode - reading folder names from {source_folder}")
        return read_folders(source_folder)


def get_player_mapping(
    source_data: list[Player], destination_data: list[Player]
) -> list[PlayerMapping]:
    logger.info("Starting player mapping process")
    player_mapping = []

    # Pre-normalize all destination names once
    normalized_to_original = {normalize(item.name): item for item in destination_data}
    available_normalized = set(normalized_to_original.keys())

    logger.info(f"Pre-normalized {len(normalized_to_original)} destination players")

    for player in source_data:
        if len(available_normalized) == 0:
            break
        candidate_normalized = get_best_match(player.name, available_normalized)
        if candidate_normalized:
            # Map back to original name
            candidate_original = normalized_to_original[candidate_normalized]
            player_mapping.append(PlayerMapping(player.id, candidate_original.id))
            # Remove matched name from available pool
            available_normalized.discard(candidate_normalized)
        else:
            logger.debug(f"No match found for player: {player.name}")

    logger.info(f"Successfully mapped {len(player_mapping)} players")
    return player_mapping


def update_faces_structure(
    src_folder_path: str, dest_folder_path: str, mapping: list[PlayerMapping]
):
    logger.info(
        f"Updating faces structure from {src_folder_path} to {dest_folder_path}"
    )
    processed = 0

    for item in mapping:
        # src_path = f"{src_folder_path}/{FACE_PATH}/{item.src_player_id}"
        src_path = f"{src_folder_path}/{item.src_player_id}"
        # dest_path = f"{dest_folder_path}/{FACE_PATH}/{item.dest_player_id}"
        dest_path = f"{dest_folder_path}/{item.dest_player_id}"

        if not os.path.exists(src_path):
            logger.debug(f"Source path does not exist: {src_path}")
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
    """Find best matching name from pre-normalized candidates. Returns normalized match or None."""
    # Normalize target once
    target_normalized = normalize(target_name)

    # Quick exact match
    if target_normalized in candidate_names:
        return target_normalized

    target_parts = target_normalized.split()
    target_surname = target_parts[-1] if target_parts else ""

    best_match = None
    best_score = -1

    for candidate_normalized in candidate_names:
        # Quick filters on normalized strings
        if abs(len(candidate_normalized) - len(target_normalized)) > 10:
            continue

        # Check if surname starts with same character
        if target_surname:
            candidate_parts = candidate_normalized.split()
            if not candidate_parts or not candidate_parts[-1].startswith(
                target_surname[0]
            ):
                continue

        # Calculate score using pre-normalized candidate
        score = calculate_name_match_score(
            target_parts, target_normalized, candidate_normalized
        )
        if score is not None and score > best_score:
            best_score = score
            best_match = candidate_normalized

    return best_match


def calculate_name_match_score(
    target_parts: list[str], target_normalized: str, candidate_normalized: str
):
    """Calculate match score using pre-normalized strings. Returns None if no match."""
    # Quick exact match check
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
    args = parse_arguments()
    logger.info("=== Starting Player Face Mapping Tool ===")

    try:
        source_data = get_source_data(args.source_csv, args.source_folder)
        destination_data = read_csv(args.destination_csv)
        mapping = get_player_mapping(source_data, destination_data)
        update_faces_structure(args.source_folder, args.dest_folder, mapping)
        logger.info("=== Finished processing successfully ===")
    except Exception as e:
        logger.critical(f"Fatal error during processing: {e}", exc_info=True)
        raise
