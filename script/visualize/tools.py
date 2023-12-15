import math
import numpy as np
import IP2Location
from IP2Location.database import IP2LocationRecord
from matplotlib import pyplot as plt
import pandas as pd
from tqdm import tqdm
import cv2 as cv

ipTool = IP2Location.IP2LocationIPTools()
database = None

IMAGE_SIZE = np.array([2060, 2068])
MAP_SIZE = np.array([2053, 2046])
BORDER_SIZE = np.array([9, 8])


def init_ip_database(path):
    """load ip database

        database can be downloaded from: 
        https://lite.ip2location.com/database/ip-country-region-city-latitude-longitude
        
        Parameters
        ----------
        path
            the path of ip database
    """
    global database
    database = IP2Location.IP2Location(path)


def get_ip_database():
    """return ip database object

    Returns
    -------
        ip database object
    """
    return database


def set_image_size(image_size, map_size, border_size):
    """set global size

    Parameters
    ----------
    image_size
        the size of the full image
    map_size
        the size of the map(without border)
    border_size
        the coordinate of the left-top corner of the map(with border)
    """
    global IMAGE_SIZE, MAP_SIZE, BORDER_SIZE
    IMAGE_SIZE = np.array(image_size)
    MAP_SIZE = np.array(map_size)
    BORDER_SIZE = np.array(border_size)


def parse_ip(ip):
    """parse ip to (latitude, longitude, country), return None if not found or latitude and longitude is 0

    Parameters
    ----------
    ip
        ip address

    Returns
    -------
        (latitude, longitude, country) or None
    """
    try:
        global database
        if database is None:
            raise Exception("IP database is not initialized")
        result: IP2LocationRecord = database.get_all(ip)
        lat, lon, coun = result.latitude, result.longitude, result.country_long
        if lat == "0.000000" and lon == "0.000000":
            return None
        return lat, lon, coun
    except Exception as e:
        print("Error:", e)
        return None


def get_mercator_value(lat, lon):
    """compute the scaled value of latitude and longitude using mercator projection

    Parameters
    ----------
    lat
        latitude
    lon
        longitude

    Returns
    -------
        (x, y) x and y is whithin [0, 1]
    """
    x = (lon + 180) / 360
    sin_latitude = math.sin(lat * math.pi / 180)
    y = 0.5 - math.log((1 + sin_latitude) / (1 - sin_latitude)) / (4 * math.pi)
    return x, y


def get_image_coordinate(lat, lon):
    """compute the actual coordinates of latitude and longitude in the image using mercator projection
    if use other Mercator Map Image, you should use set_image_size to change the size

    Parameters
    ----------
    lat
        latitude
    lon
        longitude

    Returns
    -------
        (x, y) the coordinate of the point in the image
    """
    x, y = get_mercator_value(lat, lon)
    _coord = np.array([x, y]) * MAP_SIZE + BORDER_SIZE
    return _coord.astype(int)


def get_circle_radius(cur_num, max_num, base_radius=8, max_radius=30):
    assert base_radius < max_radius
    max_radius -= base_radius
    return int(base_radius + max_radius * cur_num / max_num)


def get_circle_alpha(cur_num, max_num, base_alpha=0.5, max_alpha=1.0):
    return base_alpha + (max_alpha - base_alpha) * cur_num / max_num


def generate_ip_map(
    path: str,
    output_path="ip_map.png",
    map_dir="./Web_maps_Mercator_projection_SW.jpg",
    base_radius=8,
    max_radius=30,
    base_alpha=0.3,
    max_alpha=0.7,
):
    """generate a world map with circle representing the query frequency of each location

    Parameters
    ----------
    path
        the path of ssh log (must be json format)
    output_path, optional
        the path to save the image, by default "ip_map.png"
    map_dir, optional
        the path of the map image, by default "./Web_maps_Mercator_projection_SW.jpg"
    base_radius, optional
        the smallest radius of the circle, by default 8
    max_radius, optional
        the largest radius of the circle, by default 30
    base_alpha, optional
        the minimum alpha of the circle, by default 0.3
    max_alpha, optional
        the maximum alpha of the circle, by default 0.7

    Returns
    -------
        the image in numpy array, BGR format
    """
    print("computing data...")
    data = read_and_process_log(path)
    ip_data = data.ip
    res = ip_data.apply(
        lambda x: pd.Series(parse_ip(x), index=["latitude", "longitude", "country"])
    )  # parse_ip to location
    res.dropna(inplace=True)  # drop nan
    res = res.astype({"latitude": np.float64, "longitude": np.float64})
    res[["longitude", "latitude"]] = res.apply(
        lambda x: get_image_coordinate(x["latitude"], x["longitude"]),
        axis=1,
        result_type="expand",
    )
    res.rename(columns={"latitude": "y", "longitude": "x"}, inplace=True)

    position_list = res.value_counts(["y", "x"]).reset_index()
    position_num, max_count = position_list.shape[0], position_list["count"].max()
    colors = [
        cv.cvtColor(np.uint8([[[(i / position_num) * 180, 255, 255]]]), cv.COLOR_HSV2BGR)[0][0].tolist()
        for i in range(position_num)
    ]  # type: ignore

    # 读取图片
    img = cv.imread(map_dir)
    # 画点
    for index, row in tqdm(position_list.iterrows(), total=position_num):
        coord = (row["x"], row["y"])
        radius = get_circle_radius(row["count"], max_count, base_radius=base_radius, max_radius=max_radius)
        alpha = get_circle_alpha(row["count"], max_count, base_alpha=base_alpha, max_alpha=max_alpha)
        draw_circle(img, coord, radius, alpha, colors[index], -1)  # type: ignore
    # 保存图片
    cv.imwrite(output_path, img)
    return img


def draw_circle(
    image,
    positions: tuple[int, int],
    radius: int = 10,
    alpha=0.4,
    color=(0, 0, 255),
    thickness=-1,
    inplace=True,
):
    """draw a circle on the image

    Parameters
    ----------
    image
        the image to draw
    positions
        the coordinate of the center of the circle
    radius, optional
        the radius of the circle, by default 10
    alpha, optional
        the transparency of the circle, by default 0.4
    color, optional
        the color of the circle, by default (0, 0, 255)
    thickness, optional
        the thickness of the circle, by default -1
    inplace, optional
        whether to draw on the original image, by default True

    Returns
    -------
        the image with circle
    """
    overlay = image.copy()
    if inplace:
        output = image
    else:
        output = image.copy()
    cv.circle(overlay, positions, radius=radius, color=color, thickness=thickness)
    return cv.addWeighted(overlay, alpha, output, 1 - alpha, 0, output)


def draw_table(_ax, data, title=None):
    ax = plt.subplot(_ax)
    if title:
        ax.set_title(title)
    ax.axis("off")
    ax.axis("tight")
    table = ax.table(cellText=data.values, colLabels=data.columns, loc="center")
    table.auto_set_font_size(False)
    table.set_fontsize(10)
    table.scale(1, 1.5)
    table.auto_set_column_width(col=list(range(len(data.columns))))


def read_and_process_log(path, time_zone="Asia/Shanghai"):
    data = pd.read_json(
        path,
        lines=True,
        convert_dates=["time"],
        dtype={"port": int},
    )
    # filter non-ssh log
    data = data[data.type == "ssh"]
    data["time"] = data["time"].dt.tz_convert(time_zone)

    # split ip and port
    need_split_idx = data["port"].isna()
    need_split = data[data["port"].isna()]
    data.loc[need_split_idx, "port"] = need_split.ip.apply(lambda x: x.split(":")[1])
    data.loc[need_split_idx, "ip"] = need_split.ip.apply(lambda x: x.split(":")[0])

    return data


def auto_visualize(
    json_path,
    show_len=10,
    show_title: bool | tuple[str, str, str] = True,
    tz="Asia/Shanghai",
    save_path=None,
):
    """Auto Visualize json data from ssh log

    Parameters
    ----------
    json_path
        the json file path

    show_len, optional
        the number of "IP" and "username:password" to show, by default 10

    show_title, optional
        title for each plot, set Tuple[str, str, str] to customize,
        set False to disable, by default True

    tz, optional
        the Timezone used when converting time, by default "Asia/Shanghai"

    save_path, optional
        the path to save the image, by default None
    """
    data = read_and_process_log(json_path, tz)

    login_data = data[["user", "password"]]
    user_pwd = login_data.value_counts().head(show_len).reset_index()
    ip_list = data["ip"].value_counts().head(show_len).reset_index()
    time_list = data["time"].value_counts().resample("D").sum().reset_index()

    plt.figure(figsize=(10, 5 + show_len * 0.5))
    if show_title is True:
        title = "TOP remote IP", "TOP login username and password", "login frequency"
    elif isinstance(show_title, tuple):
        title = show_title
    else:
        title = None, None, ""
    draw_table(221, ip_list, title[0])
    draw_table(222, user_pwd, title[1])
    ax3 = plt.subplot(212)
    ax3.set_title(title[2])
    ax3.plot(time_list.time, time_list["count"], marker="o", linestyle="--", color="b")
    plt.xticks(rotation=45)

    if save_path:
        plt.savefig(save_path)
        print(f"Save image to {save_path}")
    plt.show()
