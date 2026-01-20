def calculate_stats(data):
    total = sum(data)
    count = len(data)
    average = total / count if count > 0 else 0
    return {
        "total": total,
        "count": count,
        "average": average
    }

# The last statement is the return value
result = calculate_stats(input_data)
