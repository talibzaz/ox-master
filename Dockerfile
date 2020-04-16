FROM iron/base

ENV HTTP_HOST :80
ENV GRPC_ADDR :9001
ENV ARANGO_HOST http://139.59.85.55:8529
ENV ARANGO_DB eventackle
ENV ARANGO_USERNAME root
ENV STRIPE_API_KEY sk_test_dY9jtRJdquA16pkb29JotJls
ENV ARANGO_PASSWORD qF3mKQcu7zyzBYly
ENV TICKETS_COLLECTION  tickets
ENV CARD_COLLECTION   card_details
ENV EVENTS_COLLECTION  events
ENV INJUN_GRPC_ADDR  139.59.40.163:34567
ENV EVENT_URL https://eventackle.com/search/view?id=
ENV CONFIRMATION_URL https://eventackle.com/confirm-attendee/
ENV MINIO https://minio.eventackle.com

EXPOSE 80
EXPOSE 9001

ADD ox /
CMD ["./ox", "start"]