import csv
import datetime

import redis


class ItemStat:
    def __init__(self, key_prefix, period_start, period):
        self.period_start = period_start
        self.period = datetime.timedelta(seconds=period)
        self.key_prefix = key_prefix
        self.rds = redis.StrictRedis(host="localhost", port=6379, db=0, password="test")

    def fetch_redis_data(self, key, line_limit=40):
        data = {}
        for k, v in self.rds.hgetall(key).items():
            # exp_tag:reason:index:item_id:uuid
            # c:normal:9:5120:e0286130
            exp_tag, reason, index, item_id, _ = k.decode().split(":")
            index = int(index)
            data.setdefault(exp_tag, [{} for _ in range(line_limit)])
            while index >= len(data[exp_tag]):
                data[exp_tag].append({})
            data[exp_tag][index][(item_id, reason)] = data[exp_tag][index].get((item_id, reason), 0) + int(v)
        return data

    def uniform_item_stat(self, item_dist_dict):
        if not item_dist_dict:
            return None
        key_count_list = list(item_dist_dict.items())
        total = sum([x[1] for x in key_count_list])
        for i, x in enumerate(key_count_list):
            key_count_list[i] = (x[0], x[1], x[1] / total)
        return sorted(key_count_list, key=lambda x: x[1], reverse=True)

    def stat_single_exp_data(self, dist_list):
        max_field = max(len(x) for x in dist_list)
        content_lines = []
        for i, dist in enumerate(dist_list):
            p = self.uniform_item_stat(dist)
            if p:
                item_line, prob_line = self.serialize_line(p, max_field)
                content_lines.append([f"第{i + 1}位"] + item_line)
                content_lines.append(["比例"] + prob_line)
        return content_lines

    def serialize_line(self, p, n):
        item_line, prob_line = [], []
        for i in range(min(n, len(p))):
            x = p[i]
            item_key = f"{x[0][0]},{x[0][1]}"
            prob = f"{x[2]:.2f}"
            if prob != "0.00":
                item_line.append(item_key)
                prob_line.append(prob)
        return item_line, prob_line

    def key_of_spec_time(self, time_):
        period_start = self.period_start
        period = self.period
        elapsed = time_ - period_start
        truncated_elapsed = elapsed - (elapsed % period)
        key = f"{self.key_prefix}:" + (period_start + truncated_elapsed).strftime("%Y%m%d%H%M%S")
        return key

    def stat(self, time_):
        # 转换时间格式
        data = self.fetch_redis_data(self.key_of_spec_time(time_))
        content_lines = []
        for exp_tag, dist_list in data.items():
            # 按照不同的实验统计每位置的道具分布
            line_count = sum(sum(x.values()) for x in dist_list)
            content_lines.append([f"exp_tag: {exp_tag}", f"日志数量: {line_count}"])
            content_lines.extend(self.stat_single_exp_data(dist_list))
            content_lines.append([])
        return content_lines


if __name__ == '__main__':
    now_in_hour = datetime.datetime.now().replace(minute=0, second=0, microsecond=0)
    item_stat = ItemStat("test", now_in_hour, 3600)
    lines = item_stat.stat(datetime.datetime.now())
    with open("output.csv", mode="w", newline="") as file:
        writer = csv.writer(file)
        writer.writerows(lines)
