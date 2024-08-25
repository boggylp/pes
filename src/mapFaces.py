import csv
import os
import shutil
from typing import Dict


class PlayerMapping:
    src_player_id: str
    dest_player_id: str

    def __init__(self, src_player_id: str, dest_player_id: str):
        self.src_player_id = src_player_id
        self.dest_player_id = dest_player_id


PlayerMappingDict = Dict[str, PlayerMapping]


def remove_whitespace_and_lower(text: str) -> str:
    return text.replace(" ", "").lower()


def read_csv(file_path: str):
    data = {}
    with open(file_path, mode="r", encoding="utf-8-sig") as file:
        reader = csv.DictReader(file, delimiter=";")
        for row in reader:
            player_id = row["Id"]
            player_name = row["Name"]
            data[remove_whitespace_and_lower(player_name)] = player_id
    return data


def get_player_mapping(
    source_csv: str, destination_csv: str, folder_path
) -> PlayerMappingDict:
    source_data = read_csv(source_csv)
    destination_data = read_csv(destination_csv)
    combined_dict = {}

    for player_name in source_data.keys():
        if player_name in destination_data.keys():
            combined_dict[player_name] = PlayerMapping(
                source_data[player_name], destination_data[player_name]
            )
    return combined_dict


def update_folders(
    src_folder_path: str, dest_folder_path: str, mapping: PlayerMappingDict
):
    facePath = "Asset/model/character/face/real"
    for key in mapping:
        item = mapping[key]
        src_path = f"{src_folder_path}/{facePath}/{item.src_player_id}"
        dest_path = f"{dest_folder_path}/{facePath}/{item.dest_player_id}"
        if not os.path.exists(src_path):
            continue
        if not os.path.exists(dest_path):
            os.makedirs(dest_path)
        shutil.copytree(src_path, dest_path, dirs_exist_ok=True)


if __name__ == "__main__":
    source_csv = "samples/BPB-2023-players.csv"
    destination_csv = "samples/UML Player IDs.csv"
    src_folder_path = "BPB Patch Adria Edition 2023 Faces/livecpk/Faces"
    dest_folder_path = "UML/livecpk/VRED_Faces"

    mapping = get_player_mapping(source_csv, destination_csv, src_folder_path)
    update_folders(src_folder_path, dest_folder_path, mapping)
