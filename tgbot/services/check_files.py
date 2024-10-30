import os
import re
from aiogram.types import Message
from tgbot.database.orm import AsyncORM
from tgbot.messages.texts import result_msg


async def check_file(path_to_file: str, msg: Message, lang: str):
    with open(path_to_file, "r", encoding="utf-8", errors="replace") as f:
        lines = f.readlines()

    all_len = len(lines)
    if all_len < 15000:
        return None

    unique_lines = set(lines)
    unique_len = len(unique_lines)
    dubls = all_len - unique_len

    await msg.edit_text(
        result_msg(all_len, 0, dubls, 0, lang),
    )

    added, exists = await AsyncORM.lines.insert_lines_check_existence(
        unique_lines, msg, lang, all_len, dubls
    )

    return all_len, added, dubls, exists


def format_file(path_to_file: str, order_id: int, username: str):
    with open("UrlLogPass_format.txt", "r") as f:
        url_format_list = f.read().split("\n")

    with open("LogPass_format.txt", "r") as f:
        log_format_list = f.read().split("\n")

    target_dir = os.path.join("results", f"[{order_id}]requests_@{username}")
    url_format_dir = os.path.join(target_dir, "UrlLogPass")
    log_format_dir = os.path.join(target_dir, "LogPass")

    for dir in ("results", target_dir, url_format_dir, log_format_dir):
        if not os.path.exists(dir):
            os.mkdir(dir)

    with open(path_to_file, "r", encoding="utf-8", errors="replace") as f:
        all_lines = f.readlines()

    for line in set(all_lines):
        regex = r"^(.*):(.*):(.*)$"
        try:
            line = line.replace("\x00", "")
            result = re.search(regex, line)
            if not result:
                continue

            f_url, log, f_pass = result.groups()

            # URL LOG PASS FORMAT
            for url in url_format_list:
                if url not in f_url:
                    continue

                with open(os.path.join(url_format_dir, f"{url}.txt"), "a") as f_write:
                    f_write.write(":".join([f_url, log, f_pass]) + "\n")

            # LOG PAS FORMAT
            for url in log_format_list:
                if url not in f_url:
                    continue

                with open(os.path.join(log_format_dir, f"{url}.txt"), "a") as f_write:
                    f_write.write(":".join([log, f_pass]) + "\n")

        except Exception:
            continue


async def create_file_with_searched_lines(url, save_type):
    lines = await AsyncORM.lines.get_all_lines_startswith(url)
    result = ""
    for line in lines:
        data = line.line.split(":")
        print(line, data)
        if len(data) != 4:
            continue

        https, url, log, password = data
        if save_type == "log_pass":
            result += f"{log}:{password}"
            continue
        result += f"{https}:{url}:{log}:{password}"

    url_for_name = url.split("//")[1].split("/")[0]
    name = f"({len(lines)}) {url_for_name} {save_type.replace('_', ':')}.txt"
    if not os.path.exists("temp"):
        os.mkdir("temp")

    path = os.path.join("temp", name)
    with open(path, "w") as f:
        f.write(result)

    return path
