"""
validates if songdata.json is in an expected format with the expected id sequences. 
values that can be validated will be checked as well (e.g. altTitle, chart diff, searchKeys)

the file should contain no non-ascii characters. for non-ascii characters, unicode escape sequences must be used.

every song data entry must contain the following keys: 
- id
- title
- altTitle
- artist
- charts
- searchKeys

additional validation for values: 
- id must be an integer
- title, altTitle and artist must be strings
- altTitle must be of the format '{title} ({artist})' if it is different from title.
- altTitle must only be different from title if a song with the same title already exists with a smaller id. 
- searchKeys must contain no unicode escape sequences. 

every chart entry must contain the following keys:
- id 
- diff
- level
- cc
- ver

additional validation for values: 
- id must be an integer
- diff must be either "pst", "prs", "ftr", "etr", "byd"
- level must be either "1", "2", "3", "4", "5", "6", "7", "7+", "8", "8+", "9", "9+", "10", "10+", "11", "11+", "12"
- cc must be a floating value
- ver must be a semver compatible string

song id is expected to be sorted ascendingly by the following sorting criteria: 
- version in which the song is first added
- title (converted to lowercase)
- artist (converted to lowercase)

chart id is expected to be sorted ascendingly by the following sorting criteria: 
- version in which the chart is added 
- the song's title (converted to lowercase)
- the song's artist (converted to lowercase)
- the chart's diff (PST -> PRS -> FTR -> ETR -> BYD)

skipping ids is not allowed. charts and songs should be kept in songdata.json even if they are removed from the game (e.g. Particle Arts)
"""

import json

f = open("songdata.json", "r")
s = f.read()

errs = []

# validate if the file only contains ascii characters

if not s.isascii():
    errs.append("file contains non-ascii characters")

    print(f"errors found in songdata.json:\n\n{"\n\n".join(errs)}")
    exit(1)

data = json.loads(s)

expected_song_keys = [
    "id",
    "title",
    "altTitle",
    "artist",
    "charts",
    "searchKeys",
]

expected_song_key_types = {
    "id": int,
    "title": str,
    "altTitle": str,
    "artist": str,
    "charts": list,
    "searchKeys": list,
}

expected_chart_keys = [
    "id",
    "diff",
    "level",
    "cc",
    "ver",
]

expected_chart_key_types = {
    "id": int,
    "diff": str,
    "level": str,
    "cc": float,
    "ver": str,
}

diff_ordering = {
    "pst": 0,
    "prs": 1,
    "ftr": 2,
    "etr": 3,
    "byd": 4,
}

expected_levels = [
    "1",
    "2",
    "3",
    "4",
    "5",
    "6",
    "7",
    "7+",
    "8",
    "8+",
    "9",
    "9+",
    "10",
    "10+",
    "11",
    "11+",
    "12",
]

song_dict = {}
chart_list = []

is_data_valid = True


# validate song/chart entry keys, key types and values, and prepare them for id sequence validation

for song in data:
    missing_song_keys = []

    for k in expected_song_keys:
        if k not in song:
            missing_song_keys.append(k)

    if len(missing_song_keys) > 0:
        is_data_valid = False
        errs.append(
            f"missing keys ({",".join(missing_song_keys)}) found in song entry:\n{json.dumps(song)}"
        )
        continue

    incorrect_song_key_types = []

    for k, t in expected_song_key_types.items():
        if type(song[k]) != t:
            incorrect_song_key_types.append(k)

    if len(incorrect_song_key_types) > 0:
        is_data_valid = False
        errs.append(
            f"expected type does not match for keys ({",".join(incorrect_song_key_types)}) in song entry:\n{json.dumps(song)}"
        )
        continue

    for k in song["searchKeys"]:
        if not k.isascii():
            is_data_valid = False
            errs.append(
                f"unicode sequence ({k}) found in searchKeys in song entry:\n {json.dumps(song)}"
            )
        continue

    is_charts_valid = True
    song_ver_tuple = (999, 999, 999)

    for c in song["charts"]:
        if not is_charts_valid:
            break

        missing_chart_keys = []

        for k in expected_chart_keys:
            if k not in c:
                missing_chart_keys.append(k)

        if len(missing_chart_keys) > 0:
            is_charts_valid = False
            is_data_valid = False
            errs.append(
                f"missing chart keys ({",".join(missing_chart_keys)}) found in chart entry for song '{song["title"]}':\n{json.dumps(c)}"
            )
            continue

        incorrect_chart_key_types = []

        for k, t in expected_chart_key_types.items():
            if type(c[k]) != t:
                incorrect_chart_key_types.append(k)

        if len(incorrect_chart_key_types) > 0:
            is_charts_valid = False
            is_data_valid = False
            errs.append(
                f"expected type does not match for chart keys ({",".join(incorrect_chart_key_types)}) found in chart entry for song '{song["title"]}':\n{json.dumps(c)}"
            )
            continue

        if c["diff"] not in diff_ordering:
            is_charts_valid = False
            is_data_valid = False
            errs.append(
                f"unexpected diff ({c["diff"]}) found in chart entry for song '{song["title"]}':\n{json.dumps(c)}"
            )
            continue

        if c["level"] not in expected_levels:
            is_charts_valid = False
            is_data_valid = False
            errs.append(
                f"unexpected level ({c["level"]}) found in chart entry for song '{song["title"]}':\n{json.dumps(c)}"
            )
            continue

        ver: str = c["ver"].split(".")
        is_ver_valid = True

        if len(ver) != 3:
            is_ver_valid = False

        for v in ver:
            if not v.isnumeric():
                is_ver_valid = False
                break

        if not is_ver_valid:
            is_charts_valid = False
            is_data_valid = False
            errs.append(
                f"unexpected version format ({c["ver"]}) found in chart entry for song '{song["title"]}':\n{json.dumps(c)}"
            )
            continue

        ver_tuple = (int(ver[0]), int(ver[1]), int(ver[2]))
        c["ver_tuple"] = ver_tuple
        c["title"] = song["title"]
        c["artist"] = song["artist"]
        chart_list.append(c)

        if song_ver_tuple > ver_tuple:
            song_ver_tuple = ver_tuple

    song_tuple = (song["title"], song["artist"])
    if song_tuple in song_dict:
        errs.append(
            f"duplicated song found ('{song["artist"]} - {song["title"]}'):\n{json.dumps(song)}"
        )

    song["ver_tuple"] = song_ver_tuple
    song_dict[song_tuple] = song

# basic data validation, id sequence validation and title/altTitle validation are considered as separate phases.
# the next phase should only proceed if the previous phase does not yield errors.

if not is_data_valid:
    print(f"errors found in songdata.json:\n\n{"\n\n".join(errs)}")
    exit(1)


# validate song and chart ids

chart_list.sort(
    key=lambda c: (
        c["ver_tuple"],
        c["title"].lower(),
        c["artist"].lower(),
        diff_ordering[c["diff"]],
    )
)

expected_chart_id = 1

for c in chart_list:
    title = c["title"]
    del c["title"]
    del c["artist"]
    del c["ver_tuple"]

    if c["id"] != expected_chart_id:
        is_data_valid = False
        errs.append(
            f"id for chart ({c["id"]}) does not match expected ({expected_chart_id}) for song '{title}':\n{json.dumps(c)}"
        )
        break

    expected_chart_id += 1

song_list = list(song_dict.values())
song_list.sort(key=lambda s: (s["ver_tuple"], s["title"].lower(), s["artist"].lower()))

expected_song_id = 1

for s in song_list:
    del s["ver_tuple"]

    if s["id"] != expected_song_id:
        is_data_valid = False
        errs.append(
            f"id for song ({s["id"]}) does not match expected ({expected_song_id}) for song '{s["title"]}':\n{json.dumps(s)}"
        )
        break

    expected_song_id += 1

if not is_data_valid:
    print(f"errors found in songdata.json:\n\n{"\n\n".join(errs)}")
    exit(1)


# validate title and altTitle

title_set = set()

for s in song_list:
    title = s["title"]
    altTitle = s["altTitle"]
    artist = s["artist"]

    if title not in title_set:
        if altTitle != title:
            is_data_valid = False
            errs.append(
                f"altTitle is expected to be the same as title for song '{title}':\n{json.dumps(s)}"
            )
            continue

        title_set.add(title)
    else:
        suggested_altTitle = f"{title} ({artist})"
        if altTitle != suggested_altTitle:
            is_data_valid = False
            errs.append(
                f"altTitle for this song entry is expected to be '{suggested_altTitle}':\n{json.dumps(s)}"
            )
            continue

if not is_data_valid:
    print(f"errors found in songdata.json:\n\n{"\n\n".join(errs)}")
    exit(1)

print(f"songdata.json ok")
