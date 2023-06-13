FROM mongo

RUN apt update
RUN apt install -y gawk apache2-utils netcat

COPY scripts/mongo_populator.sh /
CMD ["bash", "./mongo_populator.sh"]
