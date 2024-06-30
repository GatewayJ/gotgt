#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <string.h>

#define NUM_FILES 26214400
#define FILE_SIZE 4096
#define FILENAME_BASE "/var/tmp/tmp/file_"

pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
int fileCount = 0;

void *createFileThread(void *arg)
{
    int i = *((int *)arg);
    char filename[256];
    snprintf(filename, sizeof(filename), "%s%d", FILENAME_BASE, i);

    FILE *file = fopen(filename, "wb");
    if (file == NULL)
    {
        fprintf(stderr, "创建文件 %s 错误\n", filename);
        return NULL;
    }

    char zeroe[FILE_SIZE] = {0};
    if (fwrite(zeroe, 1, FILE_SIZE, file) != FILE_SIZE)
    {
        fprintf(stderr, "写入文件 %s 错误\n", filename);
    }
    fclose(file);

    pthread_mutex_lock(&mutex);
    fileCount++;
    if (fileCount == NUM_FILES)
    {
        printf("所有文件创建完成.\n");
    }
    pthread_mutex_unlock(&mutex);

    return NULL;
}

int main()
{
    pthread_t threads[NUM_FILES];
    int threadArgs[NUM_FILES];

    for (int i = 0; i < NUM_FILES; ++i)
    {
        threadArgs[i] = i + 1;
        if (pthread_create(&threads[i], NULL, createFileThread, &threadArgs[i]))
        {
            perror("创建线程失败");
            exit(EXIT_FAILURE);
        }
    }

    for (int i = 0; i < NUM_FILES; ++i)
    {
        pthread_join(threads[i], NULL);
    }

    return 0;
}